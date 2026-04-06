package logicoapp

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Hook func(ctx context.Context) error

type ErrorHook func(ctx context.Context, source string, err error)

type hookRegistry struct {
	mu      sync.Mutex
	onStart []Hook
	onStop  []Hook
	onError []ErrorHook
}

func (r *hookRegistry) addOnStart(h Hook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onStart = append(r.onStart, h)
}

func (r *hookRegistry) addOnStop(h Hook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onStop = append(r.onStop, h)
}

func (r *hookRegistry) addOnError(h ErrorHook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onError = append(r.onError, h)
}

func (r *hookRegistry) notifyError(ctx context.Context, source string, err error, logger Logger) {
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

func (r *hookRegistry) executeHook(ctx context.Context, h Hook, source string, logger Logger) (hookErr error) {
	defer func() {
		if rec := recover(); rec != nil {
			hookErr = fmt.Errorf("logicoapp: panic in %s hook: %v", source, rec)
			r.notifyError(ctx, "panic", hookErr, logger)
		}
	}()
	return h(ctx)
}

func (r *hookRegistry) runStartHooks(ctx context.Context, logger Logger) error {
	r.mu.Lock()
	hooks := make([]Hook, len(r.onStart))
	copy(hooks, r.onStart)
	r.mu.Unlock()

	for _, h := range hooks {
		if err := r.executeHook(ctx, h, "start", logger); err != nil {
			r.notifyError(ctx, "start", err, logger)
			return err
		}
	}
	return nil
}

func (r *hookRegistry) runStopHooks(ctx context.Context, logger Logger) error {
	r.mu.Lock()
	hooks := make([]Hook, len(r.onStop))
	copy(hooks, r.onStop)
	r.mu.Unlock()

	var errs []error
	for i := len(hooks) - 1; i >= 0; i-- {
		if err := r.executeHook(ctx, hooks[i], "stop", logger); err != nil {
			r.notifyError(ctx, "stop", err, logger)
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
