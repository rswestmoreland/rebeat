package rebeatlib

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/rswestmoreland/rebeat/config"

	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

const (
	ackframe    = "2A"
	windowsize  = "2W"
	compressed  = "2C"
	datapayload = "2D"
	jsonpayload = "2J"
	maxKeyLen   = 100 * 1024 * 1024 // 100 mb
	maxValueLen = 250 * 1024 * 1024 // 250 mb
)

type Parser struct {
        Conn               net.Conn
	config             config.Config
        readBuffer         io.Reader
        logEntriesReceived chan common.MapStr
	zlibBuffer         bytes.Buffer
	jsonBuffer         bytes.Buffer
}

// Taken from https://github.com/elasticsearch/logstash-forwarder/blob/master/event.go
type Event struct {
	Source string  `json:"source,omitempty"`
	Offset int64   `json:"offset,omitempty"`
	Line   uint64  `json:"line,omitempty"`
	Text   *string `json:"text,omitempty"`
	Fields *map[string]interface{}
}


func NewConnection(c net.Conn, logs chan common.MapStr) *Parser {
	return &Parser{
		Conn: c,
		logEntriesReceived: logs,
	}
}


// SendAck acknowledges that the payload was received successfully
func (p *Parser) SendAck(seq uint32) error {
	ackpacket := make([]byte, 6)
	copy(ackpacket[:], ackframe)
	binary.BigEndian.PutUint32(ackpacket[2:], seq)
	//logp.Info("Sending ACK for seq %d", seq)

	_, err := p.Conn.Write(ackpacket)
	if err != nil {
		return err
	}

	return nil
}

// ReadKV parses key value pairs from within the payload
func (p *Parser) ReadKV() ([]byte, []byte, error) {
	var klen, vlen uint32

	// Read key len
	binary.Read(p.readBuffer, binary.BigEndian, &klen)

	if klen > maxKeyLen {
		return nil, nil, fmt.Errorf("key exceeds max len %d, got %d bytes", maxKeyLen, klen)
	}

	// Read key
	key := make([]byte, klen)
	_, err := p.readBuffer.Read(key)
	if err != nil {
		return nil, nil, err
	}

	// Read value len
	binary.Read(p.readBuffer, binary.BigEndian, &vlen)
	if vlen > maxValueLen {
		return nil, nil, fmt.Errorf("value exceeds max len %d, got %d bytes", maxValueLen, vlen)
	}

	// Read value
	value := make([]byte, vlen)
	_, err = p.readBuffer.Read(value)
	if err != nil {
		return nil, nil, err
	}

	return key, value, nil
}

// read parses the compressed data frame
func (p *Parser) ReadPayload(wlen, plen uint32) (uint32, error) {
	var seq, count uint32
	var k, v []byte
	var err error
	j := []byte(jsonpayload)
	d := []byte(datapayload)

	p.zlibBuffer.Reset()
	io.CopyN(&p.zlibBuffer, p.Conn, int64(plen))
	zlibReader, err := zlib.NewReader(&p.zlibBuffer)
	
	if err != nil {
		logp.Err("Error initializing zlib reader")
		return seq, err
	}

	// Decompress
	p.jsonBuffer.Reset()
	io.Copy(&p.jsonBuffer, zlibReader)
	p.readBuffer = &p.jsonBuffer

	zlibReader.Close()

	b := make([]byte, 2)
	for i := uint32(0); i < wlen; i++ {
		//logp.Info("Working on wlen %d of %d", i, wlen)
		n, err := p.readBuffer.Read(b)
		if err == io.EOF {
			logp.Err("IO EOF error")
			return seq, err
		}

		if n == 0 {
			continue
		}

		if bytes.Equal(b, j) {
		// JSON data payload
			//logp.Info("Got JSON data")
			binary.Read(p.readBuffer, binary.BigEndian, &seq)
			binary.Read(p.readBuffer, binary.BigEndian, &count)
			jsonData := make([]byte, count)
			_, err := p.readBuffer.Read(jsonData)

			if err != nil {
				logp.Err("Error reading JSON data on seq %d", seq)
				return seq, err
			}

			event := common.MapStr{}

			err = json.Unmarshal(jsonData, &event)
			if err != nil {
				logp.Err("Error decoding JSON data on seq %d", seq)
				return seq, err
			}

			event["rebeat_ts"] = common.Time(time.Now())

			p.logEntriesReceived <- event

		} else if bytes.Equal(b, d) {
		// Legacy data payload
			binary.Read(p.readBuffer, binary.BigEndian, &seq)
			binary.Read(p.readBuffer, binary.BigEndian, &count)

			event := common.MapStr{}

			for j := uint32(0); j < count; j++ {
				k, v, err = p.ReadKV()
				if err != nil {
					return seq, err
				}
				event[string(k)] = string(v)
			}

			event["source"] = fmt.Sprintf("lumberjack://%s%s", event["host"], event["file"])
			event["offset"], _ = strconv.ParseInt(event["offset"].(string), 10, 64)
			event["line"] = uint64(seq)
			t := event["line"].(string)
			event["text"] = &t

			event["rebeat_ts"] = common.Time(time.Now())

			p.logEntriesReceived <- event

		} else {
			return seq, fmt.Errorf("unknown type: %s", b)
		}
	}

	return seq, nil
}

// Parse initialises the read loop and begins parsing the incoming request
func (p *Parser) Parse(timeout uint32) {
	f := make([]byte, 2)
	w := []byte(windowsize)
	c := []byte(compressed)
	z := make([]byte, 2) // empty slice is 0x0000
	var err error
	remoteHost := p.Conn.RemoteAddr().String() 

	//logp.Info("Starting Parse loop")

Read:
	for {
		// Set idle timeout, zero disables timeout, otherwise wait n seconds
		if timeout > 0 {
			p.Conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
		}

		// Expecting window length "2W"
		err = binary.Read(p.Conn, binary.BigEndian, &f)
		if err != nil {
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				logp.Warn("[%s] Timeout reading from socket", remoteHost)
			} else if err != io.EOF {
				logp.Err("[%s] Error reading %v", remoteHost, err)
			}
			break Read
		}
		//logp.Info("[%s] Got data starting with 2W, wlen %d", remoteHost, wlen)

		if !bytes.Equal(f, w) {
			// Got unexpected data
			if bytes.Equal(f, z) {
				logp.Info("[%s] Client likely closed connection", remoteHost)
			} else {
				logp.Warn("[%s] Received unknown type 0x%x, expected 2W", remoteHost, f)
			}
			break Read
		}

		var wlen uint32
		binary.Read(p.Conn, binary.BigEndian, &wlen)
		//logp.Info("[%s] Got data starting with 2W, wlen %d", remoteHost, wlen)

		// Expecting frame length "2C"
		binary.Read(p.Conn, binary.BigEndian, &f)

		if !bytes.Equal(f, c) {
			// Got unexpected data
			logp.Warn("[%s] Received unknown type 0x%x, expected 2C", remoteHost, f)
			break Read
		}

		var plen uint32
		binary.Read(p.Conn, binary.BigEndian, &plen)
		//logp.Info("[%s] Got data starting with 2C, plen %d", remoteHost, plen)
		seq, err := p.ReadPayload(wlen, plen)

		if err != nil {
			logp.Err("[%s] Error parsing %v", remoteHost, err)
			break Read
		}

		err = p.SendAck(seq)
		if err != nil {
			logp.Err("[%s] Error acking %v", remoteHost, err)
			break Read
		}

	}

}

