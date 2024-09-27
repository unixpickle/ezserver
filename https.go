package ezserver

import (
	"crypto/tls"
	"net"
	"net/http"
	"strconv"
	"sync"

	"golang.org/x/crypto/acme/autocert"
)

// HTTPS is an HTTPS server instance which can listen on one port at a time.
type HTTPS struct {
	*HTTP
	config *TLSConfig

	managerLock sync.RWMutex
	manager     *autocert.Manager
}

// NewHTTPS creates a new HTTPS server with a given handler.
// The server will not be started.
func NewHTTPS(handler http.Handler, config *TLSConfig) *HTTPS {
	return &HTTPS{HTTP: NewHTTP(handler), config: config.Clone()}
}

// SetTLSConfig sets the TLSConfig on the server.
// This may stop and restart the server.
func (h *HTTPS) SetTLSConfig(c *TLSConfig) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.config = c.Clone()
	if h.listener == nil {
		return nil
	}
	if err := h.stopInternal(); err != nil {
		return err
	}
	return h.startInternal(h.listenPort)
}

// Start runs the HTTP server on a given port.
func (h *HTTPS) Start(port int) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return h.startInternal(port)
}

// TLSConfig returns the TLSConfig for the server.
func (h *HTTPS) TLSConfig() *TLSConfig {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.config.Clone()
}

// HandleAutocertRequest handles HTTP requests to verify
// with an ACME authority that we own a domain.
// If the request is not a verification request, false is
// returned and a downstream handler should be used.
func (h *HTTPS) HandleAutocertRequest(w http.ResponseWriter, r *http.Request) bool {
	log.Printf("handling autocert request %s %s", w.Header, r.RemoteAddr)
	h.managerLock.RLock()
	manager := h.manager
	h.managerLock.RUnlock()
	if manager == nil {
		return false
	}

	handled := true
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("handling autocert request failed %s %s", w.Header, r.RemoteAddr)
		handled = false
	}
	manager.HTTPHandler(http.HandlerFunc(handler)).ServeHTTP(w, r)
	if handled {
		log.Printf("handling autocert request succeeded %s %s", w.Header, r.RemoteAddr)
	}
	return handled
}

func (h *HTTPS) startInternal(port int) error {
	if port <= 0 || port > 65535 {
		return ErrInvalidPort
	} else if h.listener != nil {
		return ErrAlreadyListening
	}

	config, manager, err := h.config.ToConfig()
	if err != nil {
		return err
	}

	if manager != nil {
		// Tip off the manager that we want to use the
		// HTTP verification method.
		manager.HTTPHandler(nil)
	}
	h.managerLock.Lock()
	h.manager = manager
	h.managerLock.Unlock()

	tcpListener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	listener := tls.NewListener(tcpListener, config)
	h.listener = &listener

	h.loopDone = make(chan struct{})
	go h.serverLoop(h.listener, h.loopDone, "https")

	h.listenPort = port

	return nil
}
