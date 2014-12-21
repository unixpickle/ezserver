package ezserver

import (
	"net"
	"net/http"
	"strconv"
	"sync"
)

// HTTP is an HTTP server instance which can listen on one port at a time.
type HTTP struct {
	mutex      sync.RWMutex
	handler    http.Handler
	listener   *net.Listener
	listenPort int
	loopDone   chan struct{}
}

// NewHTTP creates a new HTTP server with a given handler.
// The server will not be started.
func NewHTTP(handler http.Handler) *HTTP {
	return &HTTP{sync.RWMutex{}, handler, nil, 0, nil}
}

// Start runs the HTTP server on a given port.
func (self *HTTP) Start(port int) error {
	if port < 0 || port > 65535 {
		return ErrInvalidPort
	}

	self.mutex.Lock()
	defer self.mutex.Unlock()

	if self.listener != nil {
		return ErrAlreadyListening
	}

	// Create a new TCP listener
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	self.listener = &listener

	// Run the server in the background
	doneChan := make(chan struct{})
	self.loopDone = doneChan
	go func() {
		http.Serve(listener, self.handler)
		close(doneChan)
		self.mutex.Lock()
		if self.listener == &listener {
			self.listener = nil
			self.loopDone = nil
		}
		self.mutex.Unlock()
	}()

	return nil
}

// Stop stops the HTTP server.
// This method will only return once the running HTTP server has stopped
// accepting connections.
func (self *HTTP) Stop() error {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	// If not listening, return an error
	if self.listener == nil {
		return ErrNotListening
	}

	// Close the listener and nil it out
	(*self.listener).Close()
	self.listener = nil

	// Wait until the background thread's loop ends
	<-self.loopDone
	self.loopDone = nil

	self.mutex.Unlock()
	return nil
}

// Wait waits for the HTTP server to stop and then returns.
func (self *HTTP) Wait() error {
	self.mutex.Lock()

	// If not listening, return an error
	if self.listener == nil {
		self.mutex.Unlock()
		return ErrNotListening
	}

	// Get the channel and unlock the server
	ch := self.loopDone
	self.mutex.Unlock()

	// Wait for the background loop to end
	<-ch
	return nil
}
