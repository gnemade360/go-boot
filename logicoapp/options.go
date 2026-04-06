package logicoapp

import (
	"context"
	"time"

	"github.com/gnemade360/go-config/configprovider"
	"github.com/gnemade360/go-config/configutil"

	gocli "github.com/gnemade360/go-cli"
)

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

func WithConfigProvider(p configprovider.Provider) Option {
	return func(a *Application) {
		a.configProvider = p
		configHook := func(ctx context.Context) error {
			a.logger.Info("initializing configuration")
			configutil.Initialize(p)
			return nil
		}
		a.hooks.mu.Lock()
		a.hooks.onStart = append([]Hook{configHook}, a.hooks.onStart...)
		a.hooks.mu.Unlock()
	}
}

func WithCommand(cmd *gocli.Command) Option {
	return func(a *Application) {
		a.command = cmd
	}
}
