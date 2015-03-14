package ezserver

import (
	"crypto/tls"
	"net"
	"net/http"
	"strconv"
)

// HTTPS is an HTTPS server instance which can listen on one port at a time.
type HTTPS struct {
	*HTTP
	config *TLSConfig
}

// NewHTTPS creates a new HTTPS server with a given handler.
// The server will not be started.
func NewHTTPS(handler http.Handler, config *TLSConfig) *HTTPS {
	return &HTTPS{NewHTTP(handler), config.Clone()}
}

// SetTLSConfig sets the TLSConfig on the server.
// This may stop and restart the server.
func (self *HTTPS) SetTLSConfig(c *TLSConfig) error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.config = c.Clone()
	if self.listener == nil {
		return nil
	}
	if err := self.stopInternal(); err != nil {
		return err
	}
	return self.startInternal(self.listenPort)
}

// Start runs the HTTP server on a given port.
func (self *HTTPS) Start(port int) error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.startInternal(port)
}

// TLSConfig returns the TLSConfig for the server.
func (self *HTTPS) TLSConfig() *TLSConfig {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.config.Clone()
}

func (self *HTTPS) startInternal(port int) error {
	if port <= 0 || port > 65535 {
		return ErrInvalidPort
	} else if self.listener != nil {
		return ErrAlreadyListening
	}

	config, err := self.config.ToConfig()
	if err != nil {
		return err
	}

	// Create a new TCP listener
	tcpListener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	listener := tls.NewListener(tcpListener, config)
	self.listener = &listener

	// Run the server in the background
	self.loopDone = make(chan struct{})
	go self.serverLoop(self.listener, self.loopDone, "https")

	self.listenPort = port

	return nil
}
