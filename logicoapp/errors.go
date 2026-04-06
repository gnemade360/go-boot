package logicoapp

import "errors"

var (
	ErrAlreadyStarted  = errors.New("logicoapp: application already started")
	ErrNotStarted      = errors.New("logicoapp: application not started")
	ErrShutdownTimeout = errors.New("logicoapp: shutdown timed out")
	ErrStartHookFailed = errors.New("logicoapp: start hook failed")
)
