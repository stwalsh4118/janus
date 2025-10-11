package session

import (
	"io"
	"os/exec"
	"time"
)

// Manager handles session lifecycle operations
type Manager interface {
	CreateSession() (*Session, error)
	GetSession(id string) (*Session, error)
	UpdateActivity(id string) error
	UpdateProcessInfo(id string, process *exec.Cmd, stdin io.WriteCloser, stdout io.ReadCloser) error
	EndSession(id string) error
	GetAllSessions() []*Session
	CleanupInactiveSessions(timeout time.Duration)
}
