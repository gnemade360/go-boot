package logicoapp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
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
}

type Option func(*Application)

type Application struct {
	name            string
	hooks           hookRegistry
	done            chan struct{}
	once            sync.Once
	started         atomic.Bool
	shutdownTimeout time.Duration
	sigCh           chan os.Signal
}

func New(name string, opts ...Option) *Application {
	a := &Application{
		name:            name,
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

func (a *Application) Start(ctx context.Context) (<-chan struct{}, error) {
	if !a.started.CompareAndSwap(false, true) {
		return nil, ErrAlreadyStarted
	}

	log.Printf("starting application: %s", a.name)

	if err := a.hooks.runStartHooks(ctx); err != nil {
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
		log.Printf("shutting down application: %s", a.name)
		if a.sigCh != nil {
			signal.Stop(a.sigCh)
		}
		shutdownErr = a.hooks.runStopHooks(ctx)
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
