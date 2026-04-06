package logicoapp

import "time"

type Option func(*Application)

func WithLogger(l Logger) Option {
	return func(a *Application) {
		a.logger = l
	}
}

func WithShutdownTimeout(d time.Duration) Option {
	return func(a *Application) {
		a.shutdownTimeout = d
	}
}
