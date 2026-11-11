package logicoapp

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func captureStdout(t *testing.T) (read func() string, restore func()) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = w

	var mu sync.Mutex
	var buf bytes.Buffer
	copied := make(chan struct{})
	go func() {
		defer close(copied)
		chunk := make([]byte, 1024)
		for {
			n, readErr := r.Read(chunk)
			mu.Lock()
			buf.Write(chunk[:n])
			mu.Unlock()
			if readErr == io.EOF {
				return
			}
			if readErr != nil {
				return
			}
		}
	}()

	read = func() string {
		mu.Lock()
		defer mu.Unlock()
		return buf.String()
	}
	restore = func() {
		os.Stdout = original
		w.Close()
		<-copied
		r.Close()
	}
	return read, restore
}

func waitForOutput(read func() string, want string) bool {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(read(), want) {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func TestDefaultLogger_WritesFormattedInfo(t *testing.T) {
	read, restore := captureStdout(t)
	defer restore()

	logger := newDefaultLogger("svc")
	logger.Info("starting application: %s", "svc")

	if !waitForOutput(read, "starting application: svc") {
		t.Fatalf("expected formatted info message on stdout, got %q", read())
	}
}

func TestDefaultLogger_WritesError(t *testing.T) {
	read, restore := captureStdout(t)
	defer restore()

	logger := newDefaultLogger("svc")
	logger.Error("panic in error hook: %v", "boom")

	if !waitForOutput(read, "panic in error hook: boom") {
		t.Fatalf("expected formatted error message on stdout, got %q", read())
	}
}

func TestDefaultLogger_DropsDebug(t *testing.T) {
	read, restore := captureStdout(t)
	defer restore()

	logger := newDefaultLogger("svc")
	logger.Debug("debug detail")
	logger.Info("after debug")

	if !waitForOutput(read, "after debug") {
		t.Fatalf("expected info message on stdout, got %q", read())
	}
	if strings.Contains(read(), "debug detail") {
		t.Fatalf("expected debug message to be dropped, got %q", read())
	}
}
