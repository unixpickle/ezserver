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
func (h *HTTP) IsRunning() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.listener != nil
}

// AutocertHandler returns the current AutocertHandler.
//
// See SetAutocertHandler for more.
func (h *HTTP) AutocertHandler() AutocertHandler {
	h.autocertLock.RLock()
	defer h.autocertLock.RUnlock()
	return h.autocertHandler
}

// SetAutocertHandler sets the current AutocertHandler.
//
// This is set on HTTP handlers to allow them to pass
// requests to an HTTPS handler.
func (h *HTTP) SetAutocertHandler(handler AutocertHandler) {
	h.autocertLock.Lock()
	defer h.autocertLock.Unlock()
	h.autocertHandler = handler
}

// SecurityRedirects returns the list of hosts that are
// redirected to use a different amount of security.
// For HTTP servers, redirects go to HTTPS.
// For HTTPS servers, redirects go to HTTP.
//
// The returned slice is a copy of the original, so the
// caller may modify it without changing the redirects
// used by the server.
func (h *HTTP) SecurityRedirects() []string {
	h.redirectsLock.RLock()
	defer h.redirectsLock.RUnlock()
	res := make([]string, len(h.redirects))
	copy(res, h.redirects)
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
func (h *HTTP) SetSecurityRedirects(r []string) {
	h.redirectsLock.Lock()
	defer h.redirectsLock.Unlock()
	h.redirects = make([]string, len(r))
	copy(h.redirects, r)
}

// Start runs the HTTP server on a given port.
func (h *HTTP) Start(port int) error {
	if port <= 0 || port > 65535 {
		return ErrInvalidPort
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.listener != nil {
		return ErrAlreadyListening
	}

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	h.listener = &listener

	h.loopDone = make(chan struct{})
	go h.serverLoop(h.listener, h.loopDone, "http")

	h.listenPort = port

	return nil
}

// Status returns whether or not the server is running and the port on which it
// is listening (if applicable).
func (h *HTTP) Status() (bool, int) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.listener != nil, h.listenPort
}

// Stop stops the HTTP server.
// This method will only return once the running HTTP server has stopped
// accepting connections.
func (h *HTTP) Stop() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return h.stopInternal()
}

// Wait waits for the HTTP server to stop and then returns.
func (h *HTTP) Wait() error {
	h.mutex.RLock()

	if h.listener == nil {
		h.mutex.RUnlock()
		return ErrNotListening
	}

	ch := h.loopDone
	h.mutex.RUnlock()

	<-ch
	return nil
}

func (h *HTTP) serverLoop(listener *net.Listener, doneChan chan<- struct{},
	scheme string) {
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := h.AutocertHandler()
		if handler != nil && handler(w, r) {
			return
		} else if h.shouldRedirect(r.Host) {
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
		h.handler.ServeHTTP(w, r)
	})

	var server http.Server
	server.Handler = mainHandler
	server.ReadTimeout = time.Hour
	server.Serve(*listener)

	close(doneChan)
	h.mutex.Lock()
	if h.listener == listener {
		h.listener = nil
		h.loopDone = nil
	}
	h.mutex.Unlock()
}

func (h *HTTP) stopInternal() error {
	if h.listener == nil {
		return ErrNotListening
	}

	(*h.listener).Close()
	h.listener = nil

	<-h.loopDone
	h.loopDone = nil
	return nil
}

func (h *HTTP) shouldRedirect(host string) bool {
	h.redirectsLock.RLock()
	defer h.redirectsLock.RUnlock()
	for _, h := range h.redirects {
		if h == host {
			return true
		}
	}
	return false
}
