package logicoapp

import (
	"context"
	"errors"
	"testing"
)

func TestHookRegistry_StartFIFO(t *testing.T) {
	var r hookRegistry
	var order []int
	r.addOnStart(func(_ context.Context) error { order = append(order, 1); return nil })
	r.addOnStart(func(_ context.Context) error { order = append(order, 2); return nil })
	r.addOnStart(func(_ context.Context) error { order = append(order, 3); return nil })

	err := r.runStartHooks(context.Background(), defaultLogger{})
	if err != nil {
		t.Fatal(err)
	}

	expected := []int{1, 2, 3}
	for i, v := range order {
		if v != expected[i] {
			t.Fatalf("expected %v, got %v", expected, order)
		}
	}
}

func TestHookRegistry_StopLIFO(t *testing.T) {
	var r hookRegistry
	var order []int
	r.addOnStop(func(_ context.Context) error { order = append(order, 1); return nil })
	r.addOnStop(func(_ context.Context) error { order = append(order, 2); return nil })
	r.addOnStop(func(_ context.Context) error { order = append(order, 3); return nil })

	err := r.runStopHooks(context.Background(), defaultLogger{})
	if err != nil {
		t.Fatal(err)
	}

	expected := []int{3, 2, 1}
	for i, v := range order {
		if v != expected[i] {
			t.Fatalf("expected %v, got %v", expected, order)
		}
	}
}

func TestHookRegistry_StartStopsOnError(t *testing.T) {
	var r hookRegistry
	var order []int
	r.addOnStart(func(_ context.Context) error { order = append(order, 1); return nil })
	r.addOnStart(func(_ context.Context) error { return errors.New("fail") })
	r.addOnStart(func(_ context.Context) error { order = append(order, 3); return nil })

	err := r.runStartHooks(context.Background(), defaultLogger{})
	if err == nil {
		t.Fatal("expected error")
	}
	if len(order) != 1 {
		t.Fatalf("expected only first hook to run, got %v", order)
	}
}

func TestHookRegistry_StopCollectsAllErrors(t *testing.T) {
	var r hookRegistry
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	r.addOnStop(func(_ context.Context) error { return err1 })
	r.addOnStop(func(_ context.Context) error { return err2 })

	err := r.runStopHooks(context.Background(), defaultLogger{})
	if !errors.Is(err, err1) {
		t.Fatalf("expected err1 in joined error, got %v", err)
	}
	// err2 runs first (LIFO) but both should be present
	if !errors.Is(err, err2) {
		t.Fatalf("expected err2 in joined error, got %v", err)
	}
}

func TestHookRegistry_PanicRecovery_Start(t *testing.T) {
	var r hookRegistry
	r.addOnStart(func(_ context.Context) error {
		panic("boom")
	})

	err := r.runStartHooks(context.Background(), defaultLogger{})
	if err == nil {
		t.Fatal("expected error from panic recovery")
	}
	if err.Error() != "logicoapp: panic in start hook: boom" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestHookRegistry_PanicRecovery_Stop(t *testing.T) {
	var r hookRegistry
	ran := false
	r.addOnStop(func(_ context.Context) error { ran = true; return nil })
	r.addOnStop(func(_ context.Context) error { panic("boom") })

	err := r.runStopHooks(context.Background(), defaultLogger{})
	if err == nil {
		t.Fatal("expected error from panic recovery")
	}
	if !ran {
		t.Fatal("expected non-panicking stop hook to still run")
	}
}

func TestHookRegistry_ErrorNotification(t *testing.T) {
	var r hookRegistry
	var notifiedSource string
	var notifiedErr error
	r.addOnError(func(_ context.Context, source string, err error) {
		notifiedSource = source
		notifiedErr = err
	})

	hookErr := errors.New("start failed")
	r.addOnStart(func(_ context.Context) error { return hookErr })

	err := r.runStartHooks(context.Background(), defaultLogger{})
	if err == nil {
		t.Fatal("expected error")
	}
	if notifiedSource != "start" {
		t.Fatalf("expected source 'start', got %q", notifiedSource)
	}
	if notifiedErr != hookErr {
		t.Fatalf("expected notified error to be hookErr, got %v", notifiedErr)
	}
}

func TestHookRegistry_ErrorHookPanicDoesNotPropagate(t *testing.T) {
	var r hookRegistry
	r.addOnError(func(_ context.Context, _ string, _ error) {
		panic("error hook panic")
	})
	r.addOnStart(func(_ context.Context) error { return errors.New("fail") })

	err := r.runStartHooks(context.Background(), defaultLogger{})
	if err == nil {
		t.Fatal("expected error from start hook")
	}
	if err.Error() != "fail" {
		t.Fatalf("expected original error, got %v", err)
	}
}

func TestSimpleHook(t *testing.T) {
	called := false
	h := SimpleHook(func() { called = true })

	err := h(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected SimpleHook to call the function")
	}
}
