package vm

import "sync"

// StateCh wrapper for channel that can safely be closed multiple times.
type StateCh struct {
	sync.Mutex
	state  chan struct{}
	closed bool
}

// NewStateCh creates a new StateCh
func NewStateCh(closed bool) *StateCh {
	stateCh := &StateCh{state: make(chan struct{})}
	if closed {
		stateCh.Close()
	}
	return stateCh
}

// IsClosed returns either or not the channel is closed
func (s *StateCh) IsClosed() bool {
	s.Lock()
	defer s.Unlock()
	return s.closed
}

// Close the channel if not already closed
func (s *StateCh) Close() {
	s.Lock()
	defer s.Unlock()
	if !s.closed {
		close(s.state)
		s.closed = true
	}
}

// Open creates a new channel if currently closed
func (s *StateCh) Open() {
	s.Lock()
	defer s.Unlock()
	if s.closed {
		s.state = make(chan struct{})
		s.closed = false
	}
}

// Wait if channel is open, continue otherwise
func (s *StateCh) Wait() <-chan struct{} {
	return s.state
}
