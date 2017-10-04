package rebeatlib

import (
	"fmt"
	"net"
	//"crypto/tls"
	//"strings"
	//"time"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/rswestmoreland/rebeat/config"
)

type LogListener struct {
	config             config.Config
	logEntriesReceived chan common.MapStr
	logEntriesError    chan bool
}

func NewLogListener(cfg config.Config) *LogListener {
	ll := &LogListener{
		config: cfg,
	}
	return ll
}


func (ll *LogListener) Start(logEntriesReceived chan common.MapStr, logEntriesError chan bool) {

        ll.logEntriesReceived = logEntriesReceived
        ll.logEntriesError = logEntriesError

	//cert, err := tls.LoadX509KeyPair(ll.config.SSLCrt, ll.config.SSLKey)
	//if err != nil {
	//	logp.Err("Error loading keys: %v", err)
	//	ll.logEntriesError <- true
	//	return
	//}

	address := fmt.Sprintf("%s:%d", ll.config.Address, ll.config.Port)
        l, err := net.Listen("tcp", address)

        if err != nil {
                logp.Err("Error listening on tcp socket via %s: %v", address, err.Error())
                ll.logEntriesError <- true
                return
        }
        defer l.Close()

        //tlsconfig := tls.Config{Certificates: []tls.Certificate{cert}}
	//ln := tls.NewListener(l, &tlsconfig)

        logp.Info("Now listening for logs via %s on %s", ll.config.Protocol, address)

        for {
                conn, err := l.Accept()
                if err != nil {
                        logp.Err("Error accepting connection: %v", err.Error())
                        continue
                }

		go lumberConn(conn, ll.logEntriesReceived)
	}
}


// lumberConn handles an incoming connection from a lumberjack client
func lumberConn(conn net.Conn, logs chan common.MapStr) {
	//defer conn.Close()
	logp.Info("[%s] accepting lumberjack connection", conn.RemoteAddr().String())
	NewConnection(conn, logs).Parse()
	logp.Info("[%s] closing lumberjack connection", conn.RemoteAddr().String())
	conn.Close()
}


func (ll *LogListener) Shutdown() {
	close(ll.logEntriesError)
	close(ll.logEntriesReceived)
}


