# Cursor Voice Context Directory

This directory contains context files that are automatically loaded when starting a cursor-agent session. These files help cursor understand your project and recent work without requiring manual explanation.

## 📁 Directory Structure

```
.cursor-voice/
├── README.md                    # This file
├── system-prompt.md            # Voice-specific instructions (to be created)
├── project-overview.md         # Project architecture and key areas (to be created)
├── active-context.md           # Current work focus (optional, to be created)
└── conversation-summaries/     # Auto-generated summaries from past sessions
    ├── .gitkeep
    └── YYYY-MM-DD-HH-MM.md    # Generated during session end (PBI-5)
```

## 📄 Context Files

### system-prompt.md (Required)
Voice-specific instructions for cursor to optimize responses for audio consumption:
- Prefer concise answers
- Explain complex concepts simply
- Provide examples when helpful
- Remember the user is likely driving and can't read code

**Status**: To be created in PBI-3

### project-overview.md (Required)
High-level project architecture and key areas:
- Technology stack
- Main components and their responsibilities
- Important directories and files
- Common patterns and conventions

**Status**: To be created in PBI-3

### active-context.md (Optional)
What you're currently working on:
- Current PBI or feature
- Recent changes
- Files you're modifying
- Open questions or blockers

**Status**: To be created manually by developer as needed

### conversation-summaries/ (Auto-generated)
Markdown files containing summaries of past conversations:
- Generated automatically when sessions end (PBI-5)
- Named with timestamp: `YYYY-MM-DD-HH-MM.md`
- Last 2-3 summaries loaded on session start
- Provides continuity across sessions

**Status**: Will be implemented in PBI-5

## 🔄 Usage

### Automatic Loading (PBI-3)
When a new session starts, the backend will:
1. Read `system-prompt.md`
2. Read `project-overview.md`
3. Read `active-context.md` (if present)
4. Read the 2-3 most recent conversation summaries
5. Execute `git diff --name-only HEAD~3..HEAD` for recent files
6. Assemble all context into an initialization message
7. Send to cursor-agent before the first user question

### Manual Updates
You can update context files anytime:
- Edit `system-prompt.md` to adjust cursor's behavior
- Update `project-overview.md` when architecture changes
- Update `active-context.md` daily to reflect current work
- Conversation summaries are auto-generated, but you can edit them

## ⚙️ Configuration

Context loading is configured via environment variables:

```bash
CONTEXT_DIR=.cursor-voice              # Default location
MAX_CONTEXT_SUMMARIES=3                # Number of past summaries to load
GIT_RECENT_DAYS=3                      # Days of git history to include
```

## 🔒 Security Notes

- **Do not commit sensitive information** in these files
- **Do not include API keys** or credentials
- **Be mindful of proprietary code** you reference
- These files are ignored by git (except the directory structure)

## 📝 Example Files

See `docs/delivery/3/` for example context files once PBI-3 is implemented.

## 🆘 Troubleshooting

### Context not loading
- Check that required files (`system-prompt.md`, `project-overview.md`) exist
- Verify file paths are correct
- Check backend logs for context loading errors
- Ensure files are readable (permissions)

### Summaries not generating
- Verify session ends properly (not killed abruptly)
- Check `CONTEXT_DIR` environment variable
- Ensure directory has write permissions
- Review backend logs for summary generation errors

---

**Note**: This feature is fully implemented in PBI-3 (Context Initialization System).  
For PBI-0, this directory serves as a placeholder for future functionality.

