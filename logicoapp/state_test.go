package logicoapp

import (
	"sync"
	"testing"
)

func TestGlobalState_DefaultNotPaused(t *testing.T) {
	// Reset state for test
	GlobalState.Resume()
	if GlobalState.IsPaused() {
		t.Fatal("expected default state to not be paused")
	}
}

func TestGlobalState_PauseResume(t *testing.T) {
	GlobalState.Resume()

	GlobalState.Pause()
	if !GlobalState.IsPaused() {
		t.Fatal("expected state to be paused after Pause()")
	}

	GlobalState.Resume()
	if GlobalState.IsPaused() {
		t.Fatal("expected state to not be paused after Resume()")
	}
}

func TestGlobalState_Concurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)
		go func() { defer wg.Done(); GlobalState.Pause() }()
		go func() { defer wg.Done(); GlobalState.Resume() }()
		go func() { defer wg.Done(); _ = GlobalState.IsPaused() }()
	}
	wg.Wait()
}
