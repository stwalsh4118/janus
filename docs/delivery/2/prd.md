# PBI-2: Cursor-Agent Process Integration

[View in Backlog](../backlog.md#user-content-2)

## Overview

Implement integration with the cursor-agent CLI tool to enable interactive codebase queries. This includes spawning cursor-agent processes, managing stdin/stdout communication, monitoring process health, and handling process failures gracefully.

## Problem Statement

The system needs to:
- Spawn and manage cursor-agent subprocess for each session
- Communicate bidirectionally via stdin/stdout pipes
- Handle long-running responses without blocking
- Detect and recover from process crashes
- Properly terminate processes on session end
- Handle cursor-agent API rate limits and errors

## User Stories

- As a developer, I want cursor-agent to have full access to my codebase so that it can answer questions accurately
- As a developer, I want the system to recover if cursor-agent crashes so that I don't lose my session
- As a developer, I want responses even if they take time so that complex queries are answered completely
- As a system, I want to terminate cursor-agent processes cleanly so that no zombie processes remain

## Technical Approach

### Process Spawning

```go
cmd := exec.Command("cursor-agent", "chat")
cmd.Dir = codebasePath  // Working directory is codebase root

stdin, _ := cmd.StdinPipe()
stdout, _ := cmd.StdoutPipe()
stderr, _ := cmd.StderrPipe()

err := cmd.Start()
```

### Communication Protocol

1. **Sending Questions**
   - Write to stdin with newline delimiter
   - Flush buffer after each write
   - Handle write errors (broken pipe)

2. **Reading Responses**
   - Read from stdout with timeout (60s)
   - Buffer responses until complete
   - Detect end-of-response markers
   - Handle partial reads and retries

3. **Error Handling**
   - Monitor stderr for error messages
   - Detect process termination
   - Implement retry logic for transient failures

### Process Monitoring

- Background goroutine monitors `cmd.Wait()`
- Detect unexpected termination
- Log crash details to stderr
- Notify session manager of failure
- Optionally restart process automatically

### Cleanup Strategy

1. Send SIGTERM signal
2. Wait up to 5 seconds for graceful shutdown
3. Send SIGKILL if still running
4. Close all pipes
5. Reap zombie process

## UX/UI Considerations

N/A - Backend integration component. User experiences fast, reliable responses.

## Acceptance Criteria

- [ ] cursor-agent process spawns successfully with working directory set to codebase
- [ ] Questions written to stdin are received by cursor-agent
- [ ] Responses from stdout are read completely and returned to API client
- [ ] Process handle is stored in session for lifecycle management
- [ ] stderr is captured and logged for debugging
- [ ] Process crashes are detected within 1 second
- [ ] Crashed sessions return appropriate error to frontend
- [ ] Processes terminate cleanly on session end (no zombies)
- [ ] SIGKILL is sent if graceful shutdown takes >5 seconds
- [ ] Concurrent sessions have isolated cursor-agent processes
- [ ] Response timeout (60s) is enforced to prevent hanging
- [ ] Rate limit errors from Cursor API are handled gracefully
- [ ] Process health check is available via session manager

## Dependencies

- cursor-agent CLI installed and available on PATH
- Valid CURSOR_API_KEY environment variable
- Access to codebase directory
- PBI-1 (Backend Session Management) completed

## Open Questions

- What is the exact format of cursor-agent stdin/stdout? (Need to test)
- Does cursor-agent support long-running interactive sessions?
- How do we detect end-of-response in streaming output?
- What error codes does cursor-agent return?
- Are there rate limits we need to handle?
- Can we pass the API key via environment per process?

## Related Tasks

Tasks will be created when PBI moves to "Agreed" status.

