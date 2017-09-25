package rebeatlib

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/rswestmoreland/rebeat/config"
)

type LogListener struct {
	config             config.Config
	logEntriesRecieved chan common.MapStr
	logEntriesError    chan bool
}

func NewLogListener(cfg config.Config) *LogListener {
	ll := &LogListener{
		config: cfg,
	}
	return ll
}

func (ll *LogListener) Start(logEntriesRecieved chan common.MapStr, logEntriesError chan bool) {

	ll.logEntriesRecieved = logEntriesRecieved
	ll.logEntriesError = logEntriesError

	address := fmt.Sprintf("%s:%d", ll.config.Address, ll.config.Port)

	if ll.config.Protocol == "lumberjack" {
		ll.startLJ("tcp", address)
	} else {
		ll.startTCP("tcp", address)
	}

}


func (ll *LogListener) startTCP(proto string, address string) {

	l, err := net.Listen(proto, address)

	if err != nil {
		logp.Err("Error listening on % socket via %s: %v", proto, address, err.Error())
		ll.logEntriesError <- true
		return
	}
	defer l.Close()

	logp.Info("Now listening for logs via %s on %s", ll.config.Protocol, address)

	for {
		conn, err := l.Accept()
		if err != nil {
			logp.Err("Error accepting log event: %v", err.Error())
			continue
		}

		buffer := make([]byte, ll.config.MaxMsgSize)
		
		length, err := conn.Read(buffer)
		if err != nil {
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				logp.Err("Timeout reading from socket: %v", err)
				ll.logEntriesError <- true
				return
			}
		}
		go ll.processMessage(strings.TrimSpace(string(buffer[:length])))

	}
}


func (ll *LogListener) startLJ(proto string, address string) {

        l, err := net.Listen(proto, address)

        if err != nil {
                logp.Err("Error listening on % socket via %s: %v", proto, address, err.Error())
                ll.logEntriesError <- true
                return
        }
        defer l.Close()

        logp.Info("Now listening for logs via %s on %s", ll.config.Protocol, address)

        for {
                conn, err := l.Accept()
                if err != nil {
                        logp.Err("Error accepting log event: %v", err.Error())
                        continue
                }

                //buffer := make([]byte, ll.config.MaxMsgSize)
		//
                //length, err := conn.Read(buffer)
                //if err != nil {
                //        e, ok := err.(net.Error)
                //        if ok && e.Timeout() {
                //                logp.Err("Timeout reading from socket: %v", err)
                //                ll.logEntriesError <- true
                //                return
                //        }
                //}

                //go ll.processMessage(strings.TrimSpace(string(buffer[:length])))
		go lumberConn(conn)

        }
}

// lumberConn handles an incoming connection from a lumberjack client
func lumberConn(c net.Conn) {
	defer c.Close()
	logp.Info("[%s] accepting lumberjack connection", c.RemoteAddr().String())
	NewParser(c).Parse()
	logp.Info("[%s] closing lumberjack connection", c.RemoteAddr().String())
}



func (ll *LogListener) Shutdown() {
	close(ll.logEntriesError)
	close(ll.logEntriesRecieved)
}

func (ll *LogListener) processMessage(logData string) {

	if logData == "" {
		logp.Err("Event is empty")
		return
	}
	event := common.MapStr{}

	//event["message"] = input.NewParser(logData).Parse()
	//event["message"] = input.Parser(logData).Parse() <- try this next
	event["message"] = logData

	event["@timestamp"] = common.Time(time.Now())

	ll.logEntriesRecieved <- event
}

