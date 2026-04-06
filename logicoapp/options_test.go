package logicoapp

import (
	"context"
	"testing"
	"time"
)

func TestWithShutdownTimeout(t *testing.T) {
	app := New("test", WithShutdownTimeout(10*time.Second))
	if app.shutdownTimeout != 10*time.Second {
		t.Fatalf("expected 10s, got %v", app.shutdownTimeout)
	}
}

func TestWithLogger(t *testing.T) {
	logger := &testLogger{}
	app := New("test", WithLogger(logger))

	if app.logger != logger {
		t.Fatal("expected custom logger to be set")
	}
}

type testLogger struct {
	infos  []string
	errors []string
	debugs []string
}

func (l *testLogger) Info(msg string, args ...any)  { l.infos = append(l.infos, msg) }
func (l *testLogger) Error(msg string, args ...any) { l.errors = append(l.errors, msg) }
func (l *testLogger) Debug(msg string, args ...any) { l.debugs = append(l.debugs, msg) }

func TestCustomLogger_ReceivesMessages(t *testing.T) {
	logger := &testLogger{}
	app := New("test", WithLogger(logger))

	_, err := app.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	app.Shutdown(context.Background())

	if len(logger.infos) == 0 {
		t.Fatal("expected logger to receive Info messages")
	}
}
