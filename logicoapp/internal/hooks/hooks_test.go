package hooks

import (
	"context"
	"errors"
	"testing"
)

type testHookLogger struct{}

func (l *testHookLogger) Info(msg string, args ...any)  {}
func (l *testHookLogger) Error(msg string, args ...any) {}

func testLogger() HookLogger { return &testHookLogger{} }

func TestHookRegistry_StartFIFO(t *testing.T) {
	r := New()
	var order []int
	r.AddOnStart(func(_ context.Context) error { order = append(order, 1); return nil })
	r.AddOnStart(func(_ context.Context) error { order = append(order, 2); return nil })
	r.AddOnStart(func(_ context.Context) error { order = append(order, 3); return nil })

	err := r.RunStartHooks(context.Background(), testLogger())
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
	r := New()
	var order []int
	r.AddOnStop(func(_ context.Context) error { order = append(order, 1); return nil })
	r.AddOnStop(func(_ context.Context) error { order = append(order, 2); return nil })
	r.AddOnStop(func(_ context.Context) error { order = append(order, 3); return nil })

	err := r.RunStopHooks(context.Background(), testLogger())
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
	r := New()
	var order []int
	r.AddOnStart(func(_ context.Context) error { order = append(order, 1); return nil })
	r.AddOnStart(func(_ context.Context) error { return errors.New("fail") })
	r.AddOnStart(func(_ context.Context) error { order = append(order, 3); return nil })

	err := r.RunStartHooks(context.Background(), testLogger())
	if err == nil {
		t.Fatal("expected error")
	}
	if len(order) != 1 {
		t.Fatalf("expected only first hook to run, got %v", order)
	}
}

func TestHookRegistry_StopCollectsAllErrors(t *testing.T) {
	r := New()
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	r.AddOnStop(func(_ context.Context) error { return err1 })
	r.AddOnStop(func(_ context.Context) error { return err2 })

	err := r.RunStopHooks(context.Background(), testLogger())
	if !errors.Is(err, err1) {
		t.Fatalf("expected err1 in joined error, got %v", err)
	}
	if !errors.Is(err, err2) {
		t.Fatalf("expected err2 in joined error, got %v", err)
	}
}

func TestHookRegistry_PanicRecovery_Start(t *testing.T) {
	r := New()
	r.AddOnStart(func(_ context.Context) error {
		panic("boom")
	})

	err := r.RunStartHooks(context.Background(), testLogger())
	if err == nil {
		t.Fatal("expected error from panic recovery")
	}
	if err.Error() != "logicoapp: panic in start hook: boom" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestHookRegistry_PanicRecovery_Stop(t *testing.T) {
	r := New()
	ran := false
	r.AddOnStop(func(_ context.Context) error { ran = true; return nil })
	r.AddOnStop(func(_ context.Context) error { panic("boom") })

	err := r.RunStopHooks(context.Background(), testLogger())
	if err == nil {
		t.Fatal("expected error from panic recovery")
	}
	if !ran {
		t.Fatal("expected non-panicking stop hook to still run")
	}
}

func TestHookRegistry_ErrorNotification(t *testing.T) {
	r := New()
	var notifiedSource string
	var notifiedErr error
	r.AddOnError(func(_ context.Context, source string, err error) {
		notifiedSource = source
		notifiedErr = err
	})

	hookErr := errors.New("start failed")
	r.AddOnStart(func(_ context.Context) error { return hookErr })

	err := r.RunStartHooks(context.Background(), testLogger())
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
	r := New()
	r.AddOnError(func(_ context.Context, _ string, _ error) {
		panic("error hook panic")
	})
	r.AddOnStart(func(_ context.Context) error { return errors.New("fail") })

	err := r.RunStartHooks(context.Background(), testLogger())
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

func TestPrependOnStart(t *testing.T) {
	r := New()
	var order []int
	r.AddOnStart(func(_ context.Context) error { order = append(order, 2); return nil })
	r.PrependOnStart(func(_ context.Context) error { order = append(order, 1); return nil })

	err := r.RunStartHooks(context.Background(), testLogger())
	if err != nil {
		t.Fatal(err)
	}

	expected := []int{1, 2}
	for i, v := range order {
		if v != expected[i] {
			t.Fatalf("expected %v, got %v", expected, order)
		}
	}
}
