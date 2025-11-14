# Session Package

A reusable Go package for managing cursor-agent chat sessions with automatic cleanup of inactive sessions.

## Features

- **Session Management**: Create, retrieve, update, and delete sessions
- **Thread-Safe**: All operations are protected with mutexes
- **Automatic Cleanup**: Background service removes inactive sessions after a timeout
- **Cursor-Agent Integration**: Built-in support for cursor-agent command execution
- **Deep Copy Protection**: Returns clones to prevent external mutations

## Installation

```bash
go get github.com/stwalsh4118/janus/api/pkg/session
```

## Quick Start

```go
package main

import (
    "time"
    "github.com/rs/zerolog"
    "github.com/stwalsh4118/janus/api/pkg/session"
)

func main() {
    // Initialize logger (zerolog required)
    logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
    
    // Create session manager
    manager := session.NewMemorySessionManager()
    
    // Create a session
    sess, err := manager.CreateSession()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Created session: %s\n", sess.ID)
    
    // Start cleanup service (optional but recommended)
    cleanupService := session.NewCleanupService(
        manager,
        10*time.Minute,              // Session timeout
        session.DefaultCleanupInterval, // Check interval
        &logger,                     // Logger (nil uses no-op logger)
    )
    cleanupService.Start()
    defer cleanupService.Stop()
    
    // Use the session...
}
```

## Core Types

### Session

```go
type Session struct {
    ID              string
    CursorChatID    string    // Cursor-agent's internal chat session ID
    CreatedAt       time.Time
    LastActivity    time.Time
    ConversationLog []Message
}
```

### Message

```go
type Message struct {
    Role      string    // "user" or "assistant"
    Content   string
    Timestamp time.Time
}
```

## Manager Interface

The `Manager` interface provides all session operations:

```go
type Manager interface {
    CreateSession() (*Session, error)
    GetSession(id string) (*Session, error)
    UpdateActivity(id string) error
    UpdateCursorChatID(id string, cursorChatID string) error
    AskQuestion(ctx context.Context, id string, question string, workspaceDir string) (answer string, cursorChatID string, err error)
    AddToConversationLog(id string, messages []Message) error
    EndSession(id string) error
    GetAllSessions() []*Session
    CleanupInactiveSessions(timeout time.Duration)
}
```

## Implementation

### MemorySessionManager

In-memory implementation with thread-safe operations:

```go
manager := session.NewMemorySessionManager()
```

## Cleanup Service

Automatically removes sessions that have been inactive for longer than the timeout:

```go
cleanupService := session.NewCleanupService(
    manager,
    10*time.Minute,              // Sessions inactive for 10 minutes will be removed
    session.DefaultCleanupInterval, // Check every minute
    &logger,
)
cleanupService.Start()
defer cleanupService.Stop()
```

## Constants

- `DefaultCleanupInterval`: Default interval for cleanup checks (1 minute)
- `DefaultSessionTimeout`: Default session timeout (10 minutes)
- `CursorResponseTimeout`: Maximum time to wait for cursor-agent response (60 seconds)
- `HeartbeatInterval`: Expected interval between heartbeat calls (30 seconds)

## Dependencies

- `github.com/rs/zerolog` - For logging (required)
- `github.com/google/uuid` - For session ID generation

## Thread Safety

All operations are thread-safe. The `MemorySessionManager` uses `sync.RWMutex` to protect concurrent access.

## Example: Using with Cursor-Agent

```go
ctx := context.Background()
ctx, cancel := context.WithTimeout(ctx, session.CursorResponseTimeout)
defer cancel()

answer, cursorChatID, err := manager.AskQuestion(
    ctx,
    sessionID,
    "What is the meaning of life?",
    "/path/to/workspace",
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Answer: %s\n", answer)
fmt.Printf("Cursor Chat ID: %s\n", cursorChatID)
```

## License

[Your License Here]

