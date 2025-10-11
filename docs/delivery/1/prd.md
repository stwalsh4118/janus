# PBI-1: Backend Session Management

[View in Backlog](../backlog.md#user-content-1)

## Overview

Implement a Go backend server that manages cursor-agent chat sessions, providing RESTful API endpoints for session lifecycle management and query handling. The server will maintain in-memory session state and handle communication with cursor-agent processes.

## Problem Statement

Developers need a reliable backend service that can:
- Create and manage isolated chat sessions from mobile devices
- Handle concurrent requests while maintaining session isolation
- Automatically clean up inactive sessions to prevent resource leaks
- Provide health monitoring and diagnostics
- Serve the frontend application as static files

## User Stories

- As a developer, I want to start a new chat session from my phone so that I can begin asking questions about my codebase
- As a developer, I want my session to remain active while I'm using it so that conversation context is preserved
- As a developer, I want inactive sessions to be automatically cleaned up so that server resources are not wasted
- As a developer, I want to send questions and receive answers through a simple API so that the frontend can communicate with cursor

## Technical Approach

### Core Components

1. **HTTP Server** (Gin framework)
   - RESTful API endpoints
   - Static file serving for Next.js frontend
   - CORS middleware for development

2. **Session Manager**
   - In-memory session storage using map
   - Session lifecycle management
   - Automatic timeout and cleanup
   - Thread-safe operations with mutex

3. **API Endpoints**
   - `POST /api/session/start` - Create new session
   - `POST /api/ask?session_id={id}` - Send question
   - `POST /api/session/end?session_id={id}` - End session
   - `POST /api/heartbeat?session_id={id}` - Keep alive
   - `GET /api/health` - Server health check

### Data Structures

```go
type Session struct {
    ID              string
    Process         *exec.Cmd
    Stdin           io.WriteCloser
    Stdout          io.ReadCloser
    CreatedAt       time.Time
    LastActivity    time.Time
    ConversationLog []Message
}

type Message struct {
    Role      string    // "user" or "assistant"
    Content   string
    Timestamp time.Time
}
```

### Session Lifecycle

1. **Creation**: Generate UUID, spawn cursor-agent process, initialize pipes
2. **Active**: Handle queries, update LastActivity timestamp
3. **Heartbeat**: Frontend sends heartbeat every 30s
4. **Timeout**: Background goroutine checks for sessions inactive >10 minutes
5. **Cleanup**: Request summary, terminate process, remove from map

## UX/UI Considerations

N/A - Backend only component. Frontend interaction is through API.

## Acceptance Criteria

- [ ] Server starts successfully and listens on configured port
- [ ] `POST /api/session/start` creates a new session and returns session_id
- [ ] `POST /api/ask` accepts questions and returns responses from cursor-agent
- [ ] Sessions are isolated - questions in one session don't affect others
- [ ] `POST /api/heartbeat` updates LastActivity timestamp
- [ ] Background cleanup terminates sessions inactive for >10 minutes
- [ ] `POST /api/session/end` gracefully terminates session and cleans up resources
- [ ] `GET /api/health` returns server status and active session count
- [ ] Invalid session IDs return 404 with clear error message
- [ ] Server handles process crashes gracefully without crashing itself
- [ ] Request timeouts are implemented (60s for cursor responses)
- [ ] All API responses follow consistent JSON format

## Dependencies

- Go 1.21+ installed
- cursor-agent CLI available on system PATH
- Access to codebase directory for cursor-agent

## Open Questions

- What's the optimal session timeout duration for driving use case?
- Should we limit the number of concurrent sessions?
- How should we handle server restarts with active sessions?
- Do we need persistent logging of requests/responses?

## Related Tasks

See [tasks.md](./tasks.md) for the complete list of tasks associated with this PBI.

