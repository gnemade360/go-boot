package hooks

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Hook func(ctx context.Context) error

type ErrorHook func(ctx context.Context, source string, err error)

type HookLogger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

type Registry struct {
	mu      sync.Mutex
	onStart []Hook
	onStop  []Hook
	onError []ErrorHook
}

func New() *Registry {
	return &Registry{}
}

func (r *Registry) AddOnStart(h Hook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onStart = append(r.onStart, h)
}

func (r *Registry) PrependOnStart(h Hook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onStart = append([]Hook{h}, r.onStart...)
}

func (r *Registry) AddOnStop(h Hook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onStop = append(r.onStop, h)
}

func (r *Registry) AddOnError(h ErrorHook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onError = append(r.onError, h)
}

func (r *Registry) NotifyError(ctx context.Context, source string, err error, logger HookLogger) {
	r.mu.Lock()
	hooks := make([]ErrorHook, len(r.onError))
	copy(hooks, r.onError)
	r.mu.Unlock()

	for _, h := range hooks {
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic in error hook: %v", rec)
				}
			}()
			h(ctx, source, err)
		}()
	}
}

func (r *Registry) executeHook(ctx context.Context, h Hook, source string, logger HookLogger) (hookErr error) {
	defer func() {
		if rec := recover(); rec != nil {
			hookErr = fmt.Errorf("logicoapp: panic in %s hook: %v", source, rec)
			r.NotifyError(ctx, "panic", hookErr, logger)
		}
	}()
	return h(ctx)
}

func (r *Registry) RunStartHooks(ctx context.Context, logger HookLogger) error {
	r.mu.Lock()
	hooks := make([]Hook, len(r.onStart))
	copy(hooks, r.onStart)
	r.mu.Unlock()

	for _, h := range hooks {
		if err := r.executeHook(ctx, h, "start", logger); err != nil {
			r.NotifyError(ctx, "start", err, logger)
			return err
		}
	}
	return nil
}

func (r *Registry) RunStopHooks(ctx context.Context, logger HookLogger) error {
	r.mu.Lock()
	hooks := make([]Hook, len(r.onStop))
	copy(hooks, r.onStop)
	r.mu.Unlock()

	var errs []error
	for i := len(hooks) - 1; i >= 0; i-- {
		if err := r.executeHook(ctx, hooks[i], "stop", logger); err != nil {
			r.NotifyError(ctx, "stop", err, logger)
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func SimpleHook(f func()) Hook {
	return func(_ context.Context) error {
		f()
		return nil
	}
}
