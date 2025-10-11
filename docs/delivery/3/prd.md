# PBI-3: Context Initialization System

[View in Backlog](../backlog.md#user-content-3)

## Overview

Implement a system that automatically loads project context from configuration files and sends comprehensive initialization messages to cursor-agent at session start. This provides cursor with critical information about the project, current focus areas, and recent activity without requiring manual explanation from the user.

## Problem Statement

For cursor-agent to effectively answer questions while the user is driving:
- It needs to understand the project architecture and structure
- It should know what the developer is currently working on
- It should have context from recent conversations
- It must prioritize recently modified files when queries are ambiguous
- All this context must be loaded and sent automatically without user interaction

## User Stories

- As a developer, I want cursor to understand my project structure so that it can find relevant code quickly
- As a developer, I want cursor to know what I'm currently working on so that ambiguous queries default to my focus area
- As a developer, I want cursor to remember recent conversations so that I can reference past discussions
- As a developer, I want cursor to prioritize recently modified files so that questions about "the API" find my current work

## Technical Approach

### Context Files Structure

```
.cursor-voice/
├── system-prompt.md          # Required: Voice-specific instructions
├── project-overview.md        # Required: Architecture and key areas
├── active-context.md          # Optional: Current focus area
└── conversation-summaries/    # Auto-created: Past session summaries
    ├── 2024-10-05-14-30.md
    └── 2024-10-06-09-15.md
```

### Context Loading Process

1. **On Session Start**
   - Read `system-prompt.md` (required)
   - Read `project-overview.md` (required)
   - Read `active-context.md` (optional, graceful if missing)
   - List and read 2-3 most recent conversation summaries
   - Execute `git diff --name-only HEAD~3..HEAD` for recent files

2. **Context Assembly**
   - Build initialization message from template
   - Insert all loaded context
   - Format as markdown document
   - Send as first message to cursor-agent stdin

3. **Error Handling**
   - Required files missing: Return error, don't start session
   - Optional files missing: Warn in logs, continue
   - Git command fails: Log warning, omit recent files section
   - Malformed context files: Log error, use partial context

### Initialization Message Template

```markdown
[SYSTEM INITIALIZATION - Voice-based conversation with driving developer]

CRITICAL INSTRUCTIONS:
{contents of system-prompt.md}

PROJECT OVERVIEW:
{contents of project-overview.md}

CURRENT FOCUS:
{contents of active-context.md or "None specified"}

RECENT CONVERSATIONS:
{summaries from last 2-3 files, or "No previous conversations"}

RECENTLY MODIFIED FILES (last 3 days):
{git diff output, or "Unable to determine"}

---
Initialized. Ready for user questions.
```

### Configuration

```bash
CODEBASE_PATH=/path/to/code          # Required
CONTEXT_DIR=.cursor-voice             # Default
MAX_CONTEXT_SUMMARIES=3               # Default
GIT_RECENT_DAYS=3                     # Default
```

## UX/UI Considerations

- Initialization happens transparently during session start
- Frontend shows "Initializing..." status while loading context
- Initialization errors prevent session start with clear message
- Context loading should complete in <2 seconds

## Acceptance Criteria

- [ ] `.cursor-voice/` directory structure is documented
- [ ] System reads `system-prompt.md` at session start
- [ ] System reads `project-overview.md` at session start
- [ ] System reads `active-context.md` if present (optional)
- [ ] System lists conversation summary files sorted by date
- [ ] System reads the 2-3 most recent conversation summaries
- [ ] System executes git command to get recently modified files
- [ ] Initialization message is assembled from template
- [ ] All context sections are properly formatted in message
- [ ] Message is sent to cursor-agent stdin before user's first question
- [ ] Missing required files cause session start to fail with error
- [ ] Missing optional files are handled gracefully (warning logged)
- [ ] Git errors are logged but don't prevent session start
- [ ] Context loading completes within 2 seconds
- [ ] Example context files are provided in documentation

## Dependencies

- PBI-1 (Backend Session Management) completed
- PBI-2 (Cursor-Agent Integration) completed
- Git installed and available on PATH
- Codebase is a git repository

## Open Questions

- What's the optimal number of past conversations to include?
- Should we compress/summarize old conversations?
- How large can context files be before cursor performance degrades?
- Should we include git branch information?
- Do we need to handle non-git codebases?

## Related Tasks

Tasks will be created when PBI moves to "Agreed" status.

