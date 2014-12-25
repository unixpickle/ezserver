package ezserver

// Server serves some sort of web content on a specified port.
type Server interface {
	// IsRunning returns true if the server is running.
	IsRunning() bool

	// Start starts the server if it is not running.
	Start(port int) error

	// Status returns the status of the server.
	// The first return value is the result of IsRunning().
	// The second return value is the port on which the server is listenening if
	// it is running.
	Status() (bool, int)

	// Stop stops the server.
	Stop() error

	// Wait waits for the server to be stopped.
	Wait() error
}
