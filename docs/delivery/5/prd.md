# PBI-5: Conversation Summaries

[View in Backlog](../backlog.md#user-content-5)

## Overview

Implement a system that automatically generates concise summaries of each chat session when it ends, stores them as markdown files, and makes them available for context initialization in future sessions. This creates persistent memory across driving sessions.

## Problem Statement

Developers need to:
- Build continuity across multiple driving sessions
- Reference insights from past conversations without re-explaining context
- Track what topics have been discussed over time
- Provide cursor with memory of previous discussions to avoid repetition

## User Stories

- As a developer, I want summaries of my past conversations so that future sessions can reference what we discussed
- As a developer, I want summaries generated automatically so that I don't need to remember to save them
- As a developer, I want summaries to be concise so that they're useful as context without overwhelming cursor
- As a system, I want to include recent summaries in initialization so that cursor has continuity

## Technical Approach

### Summary Generation

1. **Trigger Points**
   - Manual session end via `POST /api/session/end`
   - Automatic cleanup after 10 minute timeout
   - Graceful server shutdown (SIGTERM handler)

2. **Generation Process**
   ```
   1. Send prompt to cursor-agent: 
      "Please summarize this conversation in 2-3 concise bullet points focusing on 
       key topics discussed, decisions made, and any follow-up items."
   
   2. Wait up to 10 seconds for response
   
   3. If timeout: Generate basic summary from conversation log
      - "Discussed X questions about [topics]"
   
   4. Format summary with metadata
   
   5. Save to file
   ```

3. **Summary Template**
   ```markdown
   # Conversation Summary
   
   **Date**: 2024-10-11 14:30:00
   **Duration**: 25 minutes
   **Questions Asked**: 8
   
   ## Summary
   
   - [Bullet point from cursor]
   - [Bullet point from cursor]
   - [Bullet point from cursor]
   
   ## Topics Covered
   
   - [Extracted from conversation log]
   
   ## Files Mentioned
   
   - auth/middleware.go
   - api/handlers/user.go
   ```

### File Management

- **Location**: `.cursor-voice/conversation-summaries/`
- **Naming**: `YYYY-MM-DD-HH-MM.md` (sortable by date)
- **Retention**: Keep all (user can delete manually if needed)
- **Access**: Read by context initialization system

### Integration with Context System

- PBI-3 reads 2-3 most recent summary files
- Includes in initialization message
- Provides cursor with memory of past sessions

### Error Handling

- **Cursor doesn't respond**: Generate basic summary from logs
- **Write fails**: Log error but don't fail session cleanup
- **Directory doesn't exist**: Create automatically
- **Disk full**: Log error, skip summary (non-critical)

## UX/UI Considerations

- Summary generation happens transparently
- User can review summaries manually in `.cursor-voice/conversation-summaries/`
- No UI required - backend only feature
- Frontend optionally shows "Generating summary..." during session end

## Acceptance Criteria

- [ ] Summary prompt is sent to cursor when session ends
- [ ] System waits up to 10 seconds for cursor's summary response
- [ ] Summary is formatted according to template with metadata
- [ ] Summary is saved to `.cursor-voice/conversation-summaries/YYYY-MM-DD-HH-MM.md`
- [ ] Directory is created automatically if it doesn't exist
- [ ] File naming uses ISO 8601 format with sanitized characters
- [ ] Conversation metadata (date, duration, question count) is included
- [ ] Topics and files mentioned are extracted from conversation log
- [ ] Timeout results in fallback summary from conversation log
- [ ] Write errors are logged but don't prevent session cleanup
- [ ] Summaries from timed-out sessions still attempt generation
- [ ] Summary generation doesn't delay session cleanup beyond 10s
- [ ] Generated summaries are readable and useful for future context
- [ ] System handles concurrent session endings without file conflicts

## Dependencies

- PBI-1 (Backend Session Management) completed
- PBI-2 (Cursor-Agent Integration) completed
- PBI-3 (Context Initialization System) for consumption of summaries
- File system write access to codebase directory

## Open Questions

- Should summaries include the full conversation or just highlights?
- How many past summaries should we keep before archiving?
- Should we implement summary search/indexing?
- Can we generate better summaries by analyzing conversation log directly?
- Should users be able to edit summaries before they're finalized?

## Related Tasks

Tasks will be created when PBI moves to "Agreed" status.

