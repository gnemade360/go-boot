package logicoapp

import (
	"context"
	"errors"
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

func (r *hookRegistry) runStartHooks(ctx context.Context) error {
	r.mu.Lock()
	hooks := make([]Hook, len(r.onStart))
	copy(hooks, r.onStart)
	r.mu.Unlock()

	for _, h := range hooks {
		if err := h(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *hookRegistry) runStopHooks(ctx context.Context) error {
	r.mu.Lock()
	hooks := make([]Hook, len(r.onStop))
	copy(hooks, r.onStop)
	r.mu.Unlock()

	var errs []error
	for i := len(hooks) - 1; i >= 0; i-- {
		if err := hooks[i](ctx); err != nil {
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
