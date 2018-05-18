package uploader

import (
	"os"
	"sync"
	"github.com/pborman/uuid"
)

// SessionID string
type SessionID string

// Session session to writing file
type Session struct {
	mu      *sync.Mutex
	files   map[SessionID]*os.File
	counter SessionID
}

// Add add new session
func (s *Session) Add(file *os.File) SessionID {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter = SessionID(uuid.New())
	//s.counter += 1
	s.files[s.counter] = file

	return s.counter
}

// Get get session over id
func (s *Session) Get(id SessionID) *os.File {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.files[id]
}

// Delete delete session over id
func (s *Session) Delete(id SessionID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if file, exist := s.files[id]; exist {
		file.Close()
		delete(s.files, id)
	}
}

// Len len session
func (s *Session) Len() int {
	return len(s.files)
}
