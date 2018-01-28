package ezserver

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// An AutocertHandler is a function which handles certain
// HTTP requests for ACME verification.
//
// It returns false if the request was not handled.
type AutocertHandler func(w http.ResponseWriter, r *http.Request) bool

// HTTP is an HTTP server instance which can listen on one port at a time.
type HTTP struct {
	mutex      sync.RWMutex
	handler    http.Handler
	listener   *net.Listener
	listenPort int
	loopDone   chan struct{}

	redirectsLock sync.RWMutex
	redirects     []string

	autocertLock    sync.RWMutex
	autocertHandler AutocertHandler
}

// NewHTTP creates a new HTTP server with a given handler.
// The server will not be started.
func NewHTTP(handler http.Handler) *HTTP {
	return &HTTP{handler: handler}
}

// IsRunning returns whether or not the server is accepting connections.
func (self *HTTP) IsRunning() bool {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.listener != nil
}

// AutocertHandler returns the current AutocertHandler.
//
// See SetAutocertHandler for more.
func (self *HTTP) AutocertHandler() AutocertHandler {
	self.autocertLock.RLock()
	defer self.autocertLock.RUnlock()
	return self.autocertHandler
}

// SetAutocertHandler sets the current AutocertHandler.
//
// This is set on HTTP handlers to allow them to pass
// requests to an HTTPS handler.
func (self *HTTP) SetAutocertHandler(h AutocertHandler) {
	self.autocertLock.Lock()
	defer self.autocertLock.Unlock()
	self.autocertHandler = h
}

// SecurityRedirects returns the list of hosts that are
// redirected to use a different amount of security.
// For HTTP servers, redirects go to HTTPS.
// For HTTPS servers, redirects go to HTTP.
//
// The returned slice is a copy of the original, so the
// caller may modify it without changing the redirects
// used by the server.
func (self *HTTP) SecurityRedirects() []string {
	self.redirectsLock.RLock()
	defer self.redirectsLock.RUnlock()
	res := make([]string, len(self.redirects))
	copy(res, self.redirects)
	return res
}

// SetSecurityRedirects sets the list of hosts that are
// redirected to use a different amount of security.
//
// The slice is copied before being adopted by the server,
// so the caller may modify it without changing the
// redirects used by the server.
//
// See SecurityRedirects for more.
func (self *HTTP) SetSecurityRedirects(r []string) {
	self.redirectsLock.Lock()
	defer self.redirectsLock.Unlock()
	self.redirects = make([]string, len(r))
	copy(self.redirects, r)
}

// Start runs the HTTP server on a given port.
func (self *HTTP) Start(port int) error {
	if port <= 0 || port > 65535 {
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
	self.loopDone = make(chan struct{})
	go self.serverLoop(self.listener, self.loopDone, "http")

	self.listenPort = port

	return nil
}

// Status returns whether or not the server is running and the port on which it
// is listening (if applicable).
func (self *HTTP) Status() (bool, int) {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.listener != nil, self.listenPort
}

// Stop stops the HTTP server.
// This method will only return once the running HTTP server has stopped
// accepting connections.
func (self *HTTP) Stop() error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.stopInternal()
}

// Wait waits for the HTTP server to stop and then returns.
func (self *HTTP) Wait() error {
	self.mutex.RLock()

	// If not listening, return an error
	if self.listener == nil {
		self.mutex.RUnlock()
		return ErrNotListening
	}

	// Get the channel and unlock the server
	ch := self.loopDone
	self.mutex.RUnlock()

	// Wait for the background loop to end
	<-ch
	return nil
}

func (self *HTTP) serverLoop(listener *net.Listener, doneChan chan<- struct{},
	scheme string) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := self.AutocertHandler()
		if handler != nil && handler(w, r) {
			return
		} else if self.shouldRedirect(r.Host) {
			newURL := *r.URL
			if scheme == "http" {
				newURL.Scheme = "https"
			} else {
				newURL.Scheme = "http"
			}
			newURL.Host = r.Host
			http.Redirect(w, r, newURL.String(), http.StatusTemporaryRedirect)
			return
		}
		r.URL.Scheme = scheme
		self.handler.ServeHTTP(w, r)
	})

	var server http.Server
	server.Handler = h
	server.ReadTimeout = time.Hour
	server.Serve(*listener)

	close(doneChan)
	self.mutex.Lock()
	if self.listener == listener {
		self.listener = nil
		self.loopDone = nil
	}
	self.mutex.Unlock()
}

func (self *HTTP) stopInternal() error {
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
	return nil
}

func (self *HTTP) shouldRedirect(host string) bool {
	self.redirectsLock.RLock()
	defer self.redirectsLock.RUnlock()
	for _, h := range self.redirects {
		if h == host {
			return true
		}
	}
	return false
}
