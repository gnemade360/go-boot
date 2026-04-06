package logicoapp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gnemade360/go-config/configprovider"

	gocli "github.com/gnemade360/go-cli"
)

type LogiApp interface {
	Start(ctx context.Context) (<-chan struct{}, error)
	Shutdown(ctx context.Context) error
	Run(ctx context.Context) error
	Done() <-chan struct{}
}

type LogiAppEvents interface {
	RegisterOnStart(h Hook)
	RegisterOnStop(h Hook)
	OnError(h ErrorHook)
}

type Application struct {
	name            string
	hooks           hookRegistry
	logger          Logger
	done            chan struct{}
	once            sync.Once
	started         atomic.Bool
	shutdownTimeout time.Duration
	sigCh           chan os.Signal
	configProvider  configprovider.Provider
	command         *gocli.Command
}

func New(name string, opts ...Option) *Application {
	a := &Application{
		name:            name,
		logger:          newDefaultLogger(name),
		shutdownTimeout: 30 * time.Second,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *Application) RegisterOnStart(h Hook) {
	a.hooks.addOnStart(h)
}

func (a *Application) RegisterOnStop(h Hook) {
	a.hooks.addOnStop(h)
}

func (a *Application) OnError(h ErrorHook) {
	a.hooks.addOnError(h)
}

func (a *Application) Start(ctx context.Context) (<-chan struct{}, error) {
	if !a.started.CompareAndSwap(false, true) {
		return nil, ErrAlreadyStarted
	}

	a.logger.Info("starting application: %s", a.name)

	if err := a.hooks.runStartHooks(ctx, a.logger); err != nil {
		a.started.Store(false)
		return nil, fmt.Errorf("%w: %w", ErrStartHookFailed, err)
	}

	a.done = make(chan struct{})
	a.sigCh = make(chan os.Signal, 1)
	signal.Notify(a.sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-a.sigCh:
			a.Shutdown(context.Background())
		case <-a.done:
		}
	}()

	return a.done, nil
}

func (a *Application) Shutdown(ctx context.Context) error {
	var shutdownErr error
	a.once.Do(func() {
		a.logger.Info("shutting down application: %s", a.name)
		if a.sigCh != nil {
			signal.Stop(a.sigCh)
		}
		shutdownErr = a.hooks.runStopHooks(ctx, a.logger)
		close(a.done)
	})
	if ctx.Err() != nil {
		return errors.Join(ErrShutdownTimeout, shutdownErr)
	}
	return shutdownErr
}

func (a *Application) Run(ctx context.Context) error {
	doneCh, err := a.Start(ctx)
	if err != nil {
		return err
	}

	if a.command != nil {
		errCh := make(chan error, 1)
		go func() {
			errCh <- a.command.ExecuteContext(ctx)
		}()

		select {
		case cmdErr := <-errCh:
			shutdownCtx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
			defer cancel()
			shutdownErr := a.Shutdown(shutdownCtx)
			return errors.Join(cmdErr, shutdownErr)
		case <-doneCh:
			return nil
		}
	}

	select {
	case <-ctx.Done():
	case <-doneCh:
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancel()
	return a.Shutdown(shutdownCtx)
}

func (a *Application) Done() <-chan struct{} {
	return a.done
}
