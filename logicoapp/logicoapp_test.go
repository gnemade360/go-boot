package logicoapp

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNew_Defaults(t *testing.T) {
	app := New("test-app")
	if app.name != "test-app" {
		t.Fatalf("expected name 'test-app', got %q", app.name)
	}
	if app.shutdownTimeout != 30*time.Second {
		t.Fatalf("expected 30s timeout, got %v", app.shutdownTimeout)
	}
	if app.logger == nil {
		t.Fatal("expected default logger, got nil")
	}
}

func TestNew_WithOptions(t *testing.T) {
	app := New("test", WithShutdownTimeout(5*time.Second))
	if app.shutdownTimeout != 5*time.Second {
		t.Fatalf("expected 5s timeout, got %v", app.shutdownTimeout)
	}
}

func TestStart_NoHooks(t *testing.T) {
	app := New("test")
	doneCh, err := app.Start(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer app.Shutdown(context.Background())

	select {
	case <-doneCh:
		t.Fatal("done channel should be open")
	default:
	}
}

func TestStart_HookError(t *testing.T) {
	app := New("test")
	hookErr := errors.New("init failed")
	app.RegisterOnStart(func(_ context.Context) error {
		return hookErr
	})

	_, err := app.Start(context.Background())
	if !errors.Is(err, ErrStartHookFailed) {
		t.Fatalf("expected ErrStartHookFailed, got %v", err)
	}
	if !errors.Is(err, hookErr) {
		t.Fatalf("expected wrapped hookErr, got %v", err)
	}
}

func TestStart_MultipleHooks_FIFO(t *testing.T) {
	app := New("test")
	var order []int
	app.RegisterOnStart(func(_ context.Context) error { order = append(order, 1); return nil })
	app.RegisterOnStart(func(_ context.Context) error { order = append(order, 2); return nil })
	app.RegisterOnStart(func(_ context.Context) error { order = append(order, 3); return nil })

	_, err := app.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer app.Shutdown(context.Background())

	expected := []int{1, 2, 3}
	if len(order) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, order)
	}
	for i, v := range order {
		if v != expected[i] {
			t.Fatalf("expected order %v, got %v", expected, order)
		}
	}
}

func TestStart_StopsOnFirstError(t *testing.T) {
	app := New("test")
	var order []int
	app.RegisterOnStart(func(_ context.Context) error { order = append(order, 1); return nil })
	app.RegisterOnStart(func(_ context.Context) error { return errors.New("fail") })
	app.RegisterOnStart(func(_ context.Context) error { order = append(order, 3); return nil })

	_, err := app.Start(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if len(order) != 1 || order[0] != 1 {
		t.Fatalf("expected only hook 1 to run, got %v", order)
	}
}

func TestStart_CalledTwice(t *testing.T) {
	app := New("test")
	_, err := app.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer app.Shutdown(context.Background())

	_, err = app.Start(context.Background())
	if !errors.Is(err, ErrAlreadyStarted) {
		t.Fatalf("expected ErrAlreadyStarted, got %v", err)
	}
}

func TestShutdown_Idempotent(t *testing.T) {
	app := New("test")
	callCount := 0
	app.RegisterOnStop(func(_ context.Context) error {
		callCount++
		return nil
	})

	_, err := app.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	app.Shutdown(context.Background())
	app.Shutdown(context.Background())
	app.Shutdown(context.Background())

	if callCount != 1 {
		t.Fatalf("expected stop hook called once, got %d", callCount)
	}
}

func TestShutdown_LIFO(t *testing.T) {
	app := New("test")
	var order []int
	app.RegisterOnStop(func(_ context.Context) error { order = append(order, 1); return nil })
	app.RegisterOnStop(func(_ context.Context) error { order = append(order, 2); return nil })
	app.RegisterOnStop(func(_ context.Context) error { order = append(order, 3); return nil })

	_, err := app.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	app.Shutdown(context.Background())

	expected := []int{3, 2, 1}
	if len(order) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, order)
	}
	for i, v := range order {
		if v != expected[i] {
			t.Fatalf("expected order %v, got %v", expected, order)
		}
	}
}

func TestShutdown_AllHooksRunOnError(t *testing.T) {
	app := New("test")
	var order []int
	app.RegisterOnStop(func(_ context.Context) error { order = append(order, 1); return nil })
	app.RegisterOnStop(func(_ context.Context) error { order = append(order, 2); return errors.New("fail") })
	app.RegisterOnStop(func(_ context.Context) error { order = append(order, 3); return nil })

	_, err := app.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	shutdownErr := app.Shutdown(context.Background())
	if shutdownErr == nil {
		t.Fatal("expected error from failing stop hook")
	}
	if len(order) != 3 {
		t.Fatalf("expected all 3 stop hooks to run, got %v", order)
	}
	// LIFO order
	expected := []int{3, 2, 1}
	for i, v := range order {
		if v != expected[i] {
			t.Fatalf("expected order %v, got %v", expected, order)
		}
	}
}

func TestShutdown_ContextTimeout(t *testing.T) {
	app := New("test")
	app.RegisterOnStop(func(_ context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	_, err := app.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	time.Sleep(5 * time.Millisecond) // ensure context expires

	shutdownErr := app.Shutdown(ctx)
	if !errors.Is(shutdownErr, ErrShutdownTimeout) {
		t.Fatalf("expected ErrShutdownTimeout, got %v", shutdownErr)
	}
}

func TestRun_ContextCancel(t *testing.T) {
	app := New("test")
	started := false
	stopped := false
	app.RegisterOnStart(func(_ context.Context) error { started = true; return nil })
	app.RegisterOnStop(func(_ context.Context) error { stopped = true; return nil })

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after context cancel")
	}

	if !started {
		t.Fatal("expected start hook to run")
	}
	if !stopped {
		t.Fatal("expected stop hook to run")
	}
}

func TestRun_StartError(t *testing.T) {
	app := New("test")
	app.RegisterOnStart(func(_ context.Context) error {
		return errors.New("boot failed")
	})

	err := app.Run(context.Background())
	if !errors.Is(err, ErrStartHookFailed) {
		t.Fatalf("expected ErrStartHookFailed, got %v", err)
	}
}

func TestDone_ReturnsChannel(t *testing.T) {
	app := New("test")
	_, err := app.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	doneCh := app.Done()
	select {
	case <-doneCh:
		t.Fatal("done channel should be open before shutdown")
	default:
	}

	app.Shutdown(context.Background())

	select {
	case <-doneCh:
	case <-time.After(time.Second):
		t.Fatal("done channel should be closed after shutdown")
	}
}
