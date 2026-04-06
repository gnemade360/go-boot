package logicoapp

import "sync"

var GlobalState = &globalApplicationState{mu: &sync.RWMutex{}}

type globalApplicationState struct {
	isPaused bool
	mu       *sync.RWMutex
}

func (s *globalApplicationState) IsPaused() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isPaused
}

func (s *globalApplicationState) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isPaused = true
}

func (s *globalApplicationState) Resume() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isPaused = false
}
