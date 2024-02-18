package stateCh

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
func (s *StateCh) Close() (changed bool) {
	s.Lock()
	defer s.Unlock()
	return s.close()
}

// Open creates a new channel if currently closed
func (s *StateCh) Open() (changed bool) {
	s.Lock()
	defer s.Unlock()
	return s.open()
}

// Toggle returns true if the state got opened, false otherwise
func (s *StateCh) Toggle() bool {
	s.Lock()
	defer s.Unlock()
	return s.toggle()
}

// Wait if channel is open, continue otherwise
func (s *StateCh) Wait() <-chan struct{} {
	return s.state
}

func (s *StateCh) close() (changed bool) {
	if !s.closed {
		s.closeCh()
		return true
	}
	return false
}

func (s *StateCh) open() (changed bool) {
	if s.closed {
		s.makeCh()
		return true
	}
	return false
}

func (s *StateCh) toggle() bool {
	if s.closed {
		s.makeCh()
		return true
	} else {
		s.closeCh()
		return false
	}
}

func (s *StateCh) closeCh() {
	close(s.state)
	s.closed = true
}

func (s *StateCh) makeCh() {
	s.state = make(chan struct{})
	s.closed = false
}
