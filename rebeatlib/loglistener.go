package rebeatlib

import (
	"fmt"
	"net"
	"crypto/tls"
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

	var err error
	var listener net.Listener

	address := fmt.Sprintf("%s:%d", ll.config.Address, ll.config.Port)
	var protocol string

	if ll.config.EnableSSL {
		protocol = "tcp/tls"
		var cert tls.Certificate
		cert, err = tls.LoadX509KeyPair(ll.config.SSLCrt, ll.config.SSLKey)
		if err != nil {
			logp.Err("Error loading keys: %v", err)
			ll.logEntriesError <- true
			return
		}

		tlsconfig := tls.Config{Certificates: []tls.Certificate{cert}}
		listener, err = tls.Listen("tcp", address, &tlsconfig)

	} else {
		protocol = "tcp"
		listener, err = net.Listen("tcp", address)
	}

        if err != nil {
                logp.Err("Error listening on tcp socket via %s: %v", address, err.Error())
                ll.logEntriesError <- true
                return
        }
        defer listener.Close()

        logp.Info("Now listening for logs via %s on %s", protocol, address)

        for {
                conn, err := listener.Accept()
                if err != nil {
                        logp.Err("Error accepting connection: %v", err.Error())
                        continue
                }

		go LumberConn(conn, ll.config.Timeout, ll.logEntriesReceived)
	}
}


// LumberConn handles an incoming connection from a lumberjack client
func LumberConn(conn net.Conn, timeout uint32, events chan common.MapStr) {
	defer conn.Close()
	logp.Info("[%s] Accepting lumberjack connection", conn.RemoteAddr().String())
	NewConnection(conn, events).Parse(timeout)
	logp.Info("[%s] Closing lumberjack connection", conn.RemoteAddr().String())
}


func (ll *LogListener) Shutdown() {
	close(ll.logEntriesError)
	close(ll.logEntriesReceived)
}


