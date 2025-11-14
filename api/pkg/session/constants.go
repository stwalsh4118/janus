package session

import "time"

const (
	// DefaultSessionTimeout is the duration after which inactive sessions are cleaned up
	DefaultSessionTimeout = 10 * time.Minute

	// CursorResponseTimeout is the maximum time to wait for cursor-agent response
	CursorResponseTimeout = 60 * time.Second

	// HeartbeatInterval is the expected interval between heartbeat calls
	HeartbeatInterval = 30 * time.Second
)

