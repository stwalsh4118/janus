# PBI-2: Cursor-Agent Integration Enhancements

[View in Backlog](../backlog.md#user-content-2)

## Overview

Enhance the existing command-based cursor-agent integration with robust error handling, retry logic, rate limit management, performance monitoring, and version validation. This PBI focuses on making the cursor-agent integration production-ready with comprehensive observability and reliability improvements.

## Problem Statement

The current cursor-agent integration has basic functionality but lacks production-grade features:
- No retry logic for transient failures (network issues, temporary API problems)
- Rate limits from Cursor API are not detected or handled gracefully
- Error messages are generic and don't help users understand what went wrong
- No performance monitoring or tracking of response times
- No validation of cursor-agent version compatibility
- Limited debugging capabilities when issues occur
- No configuration options for cursor-agent behavior

## User Stories

- As a developer, I want the system to automatically retry transient failures so that temporary network issues don't interrupt my conversation
- As a developer, I want clear error messages that explain what went wrong so that I know whether to retry or take action
- As a developer, I want rate limits handled gracefully so that I understand when I've hit API limits
- As a system administrator, I want to monitor cursor-agent performance so that I can identify bottlenecks and issues
- As a system administrator, I want version validation so that I know cursor-agent is compatible before the system starts

## Technical Approach

### Retry Logic with Exponential Backoff

Implement smart retry logic that distinguishes between transient and permanent failures:

```go
type RetryConfig struct {
    MaxAttempts     int           // Default: 3
    InitialDelay    time.Duration // Default: 1s
    MaxDelay        time.Duration // Default: 8s
    Multiplier      float64       // Default: 2.0
}

func (m *MemorySessionManager) AskQuestionWithRetry(
    ctx context.Context,
    sessionID string,
    question string,
    workspaceDir string,
) (string, string, error) {
    var lastErr error
    delay := m.retryConfig.InitialDelay
    
    for attempt := 1; attempt <= m.retryConfig.MaxAttempts; attempt++ {
        answer, chatID, err := m.askQuestionOnce(ctx, sessionID, question, workspaceDir)
        
        if err == nil {
            return answer, chatID, nil
        }
        
        // Don't retry on permanent failures
        if isPermanentError(err) {
            return "", "", err
        }
        
        lastErr = err
        
        // Exponential backoff before next attempt
        if attempt < m.retryConfig.MaxAttempts {
            time.Sleep(delay)
            delay = time.Duration(float64(delay) * m.retryConfig.Multiplier)
            if delay > m.retryConfig.MaxDelay {
                delay = m.retryConfig.MaxDelay
            }
        }
    }
    
    return "", "", fmt.Errorf("failed after %d attempts: %w", m.retryConfig.MaxAttempts, lastErr)
}
```

**Permanent Errors (No Retry):**
- Authentication failures
- Invalid input/malformed questions
- Rate limit errors (handled separately)
- Context timeout/cancellation

**Transient Errors (Retry):**
- Network connectivity issues
- Temporary API unavailability
- Process spawn failures
- Parse errors (might be transient corruption)

### Rate Limit Handling

Detect and handle rate limit errors from cursor-agent:

```go
type ErrorCategory int

const (
    ErrCategoryUnknown ErrorCategory = iota
    ErrCategoryAuth
    ErrCategoryRateLimit
    ErrCategoryNetwork
    ErrCategoryInvalidInput
    ErrCategoryTimeout
)

func categorizeError(stderr string, err error) ErrorCategory {
    stderrLower := strings.ToLower(stderr)
    
    // Check for rate limit indicators
    if strings.Contains(stderrLower, "rate limit") ||
       strings.Contains(stderrLower, "429") ||
       strings.Contains(stderrLower, "too many requests") {
        return ErrCategoryRateLimit
    }
    
    // Check for auth errors
    if strings.Contains(stderrLower, "unauthorized") ||
       strings.Contains(stderrLower, "invalid api key") ||
       strings.Contains(stderrLower, "401") {
        return ErrCategoryAuth
    }
    
    // Check for network errors
    if strings.Contains(stderrLower, "network") ||
       strings.Contains(stderrLower, "connection") ||
       strings.Contains(stderrLower, "timeout") {
        return ErrCategoryNetwork
    }
    
    // Check for context timeout
    if errors.Is(err, context.DeadlineExceeded) {
        return ErrCategoryTimeout
    }
    
    return ErrCategoryUnknown
}
```

**Rate Limit Response:**
- Return specific error code: `RATE_LIMIT_EXCEEDED`
- Include estimated retry time if available in stderr
- Log rate limit events with session ID and timestamp
- Frontend can display user-friendly message

### Enhanced Error Parsing and User-Friendly Messages

Map technical errors to clear, actionable messages:

```go
type CursorAgentError struct {
    Category    ErrorCategory
    Message     string // User-friendly message
    Technical   string // Technical details for logs
    Retryable   bool
    ErrorCode   string // API error code
}

func parseError(err error, stderr string) *CursorAgentError {
    category := categorizeError(stderr, err)
    
    switch category {
    case ErrCategoryAuth:
        return &CursorAgentError{
            Category:  ErrCategoryAuth,
            Message:   "Authentication failed. Please check your CURSOR_API_KEY configuration.",
            Technical: stderr,
            Retryable: false,
            ErrorCode: "AUTH_FAILED",
        }
    case ErrCategoryRateLimit:
        return &CursorAgentError{
            Category:  ErrCategoryRateLimit,
            Message:   "API rate limit exceeded. Please wait a moment before trying again.",
            Technical: stderr,
            Retryable: false,
            ErrorCode: "RATE_LIMIT_EXCEEDED",
        }
    case ErrCategoryNetwork:
        return &CursorAgentError{
            Category:  ErrCategoryNetwork,
            Message:   "Network connection issue. Retrying...",
            Technical: stderr,
            Retryable: true,
            ErrorCode: "NETWORK_ERROR",
        }
    case ErrCategoryTimeout:
        return &CursorAgentError{
            Category:  ErrCategoryTimeout,
            Message:   "Request timed out. The query may be too complex or the service is slow.",
            Technical: stderr,
            Retryable: true,
            ErrorCode: "TIMEOUT",
        }
    default:
        return &CursorAgentError{
            Category:  ErrCategoryUnknown,
            Message:   "An unexpected error occurred. Please try again.",
            Technical: stderr,
            Retryable: true,
            ErrorCode: "UNKNOWN_ERROR",
        }
    }
}
```

### Performance Metrics

Track and monitor cursor-agent performance:

```go
type SessionMetrics struct {
    RequestCount    int
    TotalDuration   time.Duration
    MinDuration     time.Duration
    MaxDuration     time.Duration
    LastRequestTime time.Time
    SlowQueryCount  int // Queries >30s
}

func (m *MemorySessionManager) trackRequestMetrics(
    sessionID string,
    duration time.Duration,
) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    session, exists := m.sessions[sessionID]
    if !exists {
        return
    }
    
    metrics := &session.Metrics
    metrics.RequestCount++
    metrics.TotalDuration += duration
    metrics.LastRequestTime = time.Now()
    
    if metrics.MinDuration == 0 || duration < metrics.MinDuration {
        metrics.MinDuration = duration
    }
    if duration > metrics.MaxDuration {
        metrics.MaxDuration = duration
    }
    if duration > 30*time.Second {
        metrics.SlowQueryCount++
        logger.Get().Warn().
            Str("session_id", sessionID).
            Dur("duration", duration).
            Msg("Slow cursor-agent query detected")
    }
}
```

**Health Endpoint Enhancement:**
```go
type HealthResponse struct {
    Status          string                 `json:"status"`
    ActiveSessions  int                    `json:"active_sessions"`
    CursorVersion   string                 `json:"cursor_agent_version"`
    Metrics         HealthMetrics          `json:"metrics"`
}

type HealthMetrics struct {
    TotalRequests   int            `json:"total_requests"`
    AvgResponseTime string         `json:"avg_response_time"`
    SlowQueries     int            `json:"slow_queries"`
}
```

### Version Checking

Validate cursor-agent version on server startup:

```go
func checkCursorAgentVersion() (string, error) {
    cmd := exec.Command("cursor-agent", "--version")
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to get cursor-agent version: %w", err)
    }
    
    version := strings.TrimSpace(string(output))
    
    // Parse and validate minimum version
    if !isCompatibleVersion(version, MIN_CURSOR_AGENT_VERSION) {
        return version, fmt.Errorf(
            "cursor-agent version %s is too old (minimum: %s)",
            version,
            MIN_CURSOR_AGENT_VERSION,
        )
    }
    
    logger.Get().Info().
        Str("version", version).
        Msg("cursor-agent version validated")
    
    return version, nil
}
```

### Enhanced Logging

Comprehensive logging for debugging and monitoring:

```go
func (m *MemorySessionManager) askQuestionOnce(...) {
    startTime := time.Now()
    
    // Log request start
    logger.Get().Debug().
        Str("session_id", sessionID).
        Str("question_preview", truncate(question, 50)).
        Strs("args", args).
        Msg("Executing cursor-agent command")
    
    // Execute command...
    
    duration := time.Since(startTime)
    
    // Always log stderr if present
    if stderr.Len() > 0 {
        logger.Get().Warn().
            Str("session_id", sessionID).
            Str("stderr", stderr.String()).
            Msg("cursor-agent stderr output")
    }
    
    // Log request completion
    logger.Get().Info().
        Str("session_id", sessionID).
        Str("cursor_chat_id", cursorChatID).
        Dur("duration", duration).
        Int("response_length", len(answer)).
        Msg("cursor-agent request completed")
}
```

### Command Configuration

Support configuration via environment variables:

```bash
# Required
CURSOR_API_KEY=<key>
WORKSPACE_DIR=/path/to/codebase

# Optional cursor-agent configuration
CURSOR_AGENT_MODEL=gpt-4          # If supported by cursor-agent
CURSOR_AGENT_TIMEOUT=60s          # Command timeout
CURSOR_AGENT_EXTRA_FLAGS=--verbose # Additional flags

# Retry configuration
CURSOR_RETRY_MAX_ATTEMPTS=3
CURSOR_RETRY_INITIAL_DELAY=1s
CURSOR_RETRY_MAX_DELAY=8s
```

**Configuration Structure:**
```go
type CursorAgentConfig struct {
    Model       string
    Timeout     time.Duration
    ExtraFlags  []string
    RetryConfig RetryConfig
}

func loadConfig() *CursorAgentConfig {
    return &CursorAgentConfig{
        Model:      getEnvOrDefault("CURSOR_AGENT_MODEL", ""),
        Timeout:    parseDuration(getEnvOrDefault("CURSOR_AGENT_TIMEOUT", "60s")),
        ExtraFlags: parseFlags(getEnv("CURSOR_AGENT_EXTRA_FLAGS")),
        RetryConfig: RetryConfig{
            MaxAttempts:  parseInt(getEnvOrDefault("CURSOR_RETRY_MAX_ATTEMPTS", "3")),
            InitialDelay: parseDuration(getEnvOrDefault("CURSOR_RETRY_INITIAL_DELAY", "1s")),
            MaxDelay:     parseDuration(getEnvOrDefault("CURSOR_RETRY_MAX_DELAY", "8s")),
            Multiplier:   2.0,
        },
    }
}
```

## UX/UI Considerations

- Users receive clear, actionable error messages instead of generic failures
- Transient failures are automatically retried without user intervention
- Rate limit errors inform users to wait before retrying
- Frontend can use error codes to customize UI messaging
- Performance monitoring helps identify and fix slow response issues

## Acceptance Criteria

- [ ] Retry logic automatically retries transient failures up to 3 times with exponential backoff
- [ ] Rate limit errors are detected from stderr and returned with error code `RATE_LIMIT_EXCEEDED`
- [ ] Authentication errors are detected and returned with error code `AUTH_FAILED`
- [ ] Network errors are detected and classified as retryable
- [ ] Timeout errors are handled separately from other failures
- [ ] No retries occur for permanent failures (auth, rate limit, invalid input)
- [ ] Response time is tracked for every cursor-agent request
- [ ] Slow queries (>30s) are logged with warning level
- [ ] Session metrics include min/max/avg response times and request count
- [ ] Health endpoint returns cursor-agent version and performance metrics
- [ ] Cursor-agent version is validated on server startup
- [ ] Server startup fails if cursor-agent version is incompatible
- [ ] Stderr output is always captured and logged
- [ ] Command arguments are logged at debug level for troubleshooting
- [ ] Configuration supports environment variables for retry behavior
- [ ] Configuration supports custom cursor-agent flags if provided
- [ ] Error responses include both user-friendly messages and error codes
- [ ] All cursor-agent errors are categorized (auth, rate limit, network, timeout, unknown)

## Dependencies

- cursor-agent CLI installed and available on PATH
- Valid CURSOR_API_KEY environment variable
- Access to codebase directory
- PBI-1 (Backend Session Management) completed

## Open Questions

- What's the minimum cursor-agent version we should require?
- What retry strategy works best for the driving use case (is 3 attempts enough)?
- Should we implement a circuit breaker pattern for repeated failures?
- What are the actual rate limits enforced by Cursor API?
- Should performance metrics be exposed to the frontend for user visibility?
- Does cursor-agent support model selection or other configuration flags?
- How should we handle rate limit backoff (fixed time vs. exponential)?

## Related Tasks

Tasks will be created when PBI moves to "Agreed" status.
