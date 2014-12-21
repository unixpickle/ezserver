package ezserver

import "errors"

// Errors relating to HTTP or HTTPS servers.
var (
	ErrAlreadyListening = errors.New("Already listening.")
	ErrInvalidPort      = errors.New("Invalid port.")
	ErrNotListening     = errors.New("Not listening.")
)
