# Cursor Voice Chat - MVP Specification

## Project Overview

A voice-enabled interface for chatting with your codebase while driving. The system allows developers to have natural, hands-free conversations about their code using Cursor's agent capabilities, without needing to look at screens or remember specific file paths.

### Core Problem Being Solved
- Developers want to think through code problems while driving
- Traditional voice assistants (Siri) cut off too early and can't understand codebases
- Referencing specific files/locations isn't practical while driving
- Need natural, conversational interface that handles vague queries like "that auth thing I was working on"

### Key Principle
The user is driving and cannot see any code. All interactions must be audio-only and conversational.

---

## Architecture Design

### High-Level Flow
```
[Phone Browser]
    ↓ (holds button, Web Speech API records)
[Next.js Frontend]
    ↓ (sends transcript via HTTP/Tailscale)
[Go Backend on Developer's Computer]
    ↓ (manages cursor-agent chat session)
[cursor-agent CLI]
    ↓ (has full codebase context via Cursor)
[Go Backend]
    ↓ (returns response)
[Next.js Frontend]
    ↓ (Web Speech API speaks response)
[User hears answer while driving]
```

### Network Architecture
- **Connection**: Tailscale VPN between phone and home computer
- **Server Location**: Developer's computer (where codebase lives)
- **Client**: Mobile browser accessing via Tailscale IP
- **No cloud services**: Everything runs locally for privacy and speed

---

## Tech Stack

### Backend
- **Language**: Go
- **Server**: Standard net/http
- **Process Management**: os/exec for cursor-agent subprocess
- **Session Storage**: In-memory map (no database)

### Frontend
- **Framework**: Next.js
- **Voice Input**: Web Speech API (webkitSpeechRecognition)
- **Voice Output**: Web Speech API (speechSynthesis)
- **Styling**: Tailwind CSS (user's preference)
- **Deployment**: Served by Go backend as static files

### External Dependencies
- **Cursor CLI**: cursor-agent (already installed on developer's machine)
- **Tailscale**: For secure phone-to-computer connection
- **Browser APIs**: Chrome/Safari on iOS/Android

---

## Core MVP Features

### 1. Session Management
- Each web app connection creates a new cursor-agent chat session
- Sessions are isolated and ephemeral
- Automatic cleanup on inactivity (10 minute timeout)
- Graceful shutdown with conversation summary generation

### 2. Context Initialization
Before the user asks any questions, the system sends an initialization message to cursor-agent containing:
- System prompt with voice-specific instructions
- Project overview from `.cursor-voice/project-overview.md`
- Current focus area from `.cursor-voice/active-context.md`
- Summaries from last 2-3 conversations
- List of recently modified files (last 3 days)

### 3. Voice Interface
- **Push-to-talk button**: Hold to record, release to send
- **Continuous recording**: No time limits, supports long rambling
- **Real-time transcript**: Shows what user said as they speak
- **Response display**: Shows cursor's answer in text
- **Text-to-speech**: Automatically reads response aloud
- **Large touch targets**: Optimized for use while driving (safety)

### 4. Conversational AI Behavior
Cursor is instructed to:
- Always state which files/code areas it's examining first
- Ask for clarification when queries are ambiguous
- List multiple relevant areas if found
- Keep responses conversational and audio-friendly
- Remember conversation context for follow-up questions
- Prioritize recently modified files when context is unclear

### 5. Conversation Summaries
At session end:
- System asks cursor to generate 2-3 bullet point summary
- Summary saved to `.cursor-voice/conversation-summaries/YYYY-MM-DD-HH-MM.md`
- Future sessions can reference past conversations
- Helps build persistent knowledge across sessions

---

## User Flow

### Starting a Session
1. User opens web app on phone (via Tailscale URL)
2. Frontend calls `POST /api/session/start`
3. Backend spawns cursor-agent process
4. Backend sends initialization message with all context
5. Backend returns session_id to frontend
6. UI shows "Ready" state with push-to-talk button

### Having a Conversation
1. User holds push-to-talk button
2. Web Speech API transcribes speech in real-time
3. User releases button when done speaking
4. Frontend sends transcript to `POST /api/ask?session_id=xyz`
5. Backend pipes question to cursor-agent via stdin
6. Backend reads response from cursor-agent stdout
7. Backend returns response to frontend
8. Frontend displays text and speaks it via TTS
9. Repeat for follow-up questions

### Ending a Session
1. User closes browser tab OR 10 minutes of inactivity
2. Frontend calls `POST /api/session/end?session_id=xyz` (if manual)
3. Backend asks cursor to summarize conversation
4. Backend saves summary to conversation-summaries/
5. Backend kills cursor-agent process
6. Backend removes session from memory

---

## Implementation Details

### Backend API Endpoints

#### `POST /api/session/start`
**Request Body**: None (or optional context hints)
**Response**: 
```json
{
  "session_id": "uuid-here",
  "status": "ready"
}
```
**Logic**:
1. Generate unique session ID
2. Read all context files from `.cursor-voice/`
3. Get list of recently modified files (git diff)
4. Spawn `cursor-agent chat` process
5. Send initialization message via stdin
6. Store session in memory map
7. Return session_id

#### `POST /api/ask?session_id={id}`
**Request Body**:
```json
{
  "question": "how does the auth middleware work?"
}
```
**Response**:
```json
{
  "answer": "I'm looking at your authentication middleware in middleware/jwt.go...",
  "timestamp": "2024-10-11T20:30:00Z"
}
```
**Logic**:
1. Validate session exists
2. Update last_activity timestamp
3. Write question to cursor stdin
4. Read response from cursor stdout (blocking)
5. Log message to conversation history
6. Return response

#### `POST /api/session/end?session_id={id}`
**Request Body**: None
**Response**:
```json
{
  "status": "ended",
  "summary_saved": true
}
```
**Logic**:
1. Send "Summarize this conversation in 3 bullet points" to cursor
2. Wait for response (with 10 second timeout)
3. Save summary to `.cursor-voice/conversation-summaries/`
4. Kill cursor process
5. Remove session from memory

#### `GET /api/health`
**Response**:
```json
{
  "status": "ok",
  "codebase_path": "/path/to/code",
  "cursor_available": true,
  "active_sessions": 1
}
```

#### `POST /api/heartbeat?session_id={id}`
**Purpose**: Keep session alive, called by frontend every 30 seconds
**Response**: `{ "status": "ok" }`

### Session Structure (Go)
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

### Process Lifecycle Management
- **Spawning**: `cursor-agent chat` with stdin/stdout pipes
- **Communication**: Write questions to stdin, read from stdout
- **Monitoring**: Track process health, restart on crash
- **Cleanup**: SIGTERM → wait 5s → SIGKILL if needed
- **Timeout**: Background goroutine checks session.LastActivity every minute

### Context Folder Structure
```
your-codebase/
├── .cursor-voice/
│   ├── system-prompt.md
│   ├── project-overview.md
│   ├── active-context.md
│   └── conversation-summaries/
│       ├── 2024-10-05-14-30.md
│       ├── 2024-10-06-09-15.md
│       └── ...
```

### Initialization Message Template
```markdown
[SYSTEM INITIALIZATION - Voice-based conversation with driving developer]

CRITICAL INSTRUCTIONS:
- User CANNOT see code or screens while driving
- ALWAYS state which files you're examining before answering
- If query is ambiguous, list all relevant options found and ask for clarification
- Keep responses conversational and concise (audio-friendly)
- Remember conversation context for follow-up questions
- Prioritize recently modified files when context is unclear
- Use natural language for locations (e.g., "in your auth handler" not "auth.go:45")

PROJECT OVERVIEW:
{contents of project-overview.md}

CURRENT FOCUS:
{contents of active-context.md}

RECENT CONVERSATIONS:
{summaries from last 2-3 conversation files}

RECENTLY MODIFIED FILES (last 3 days):
{git diff --name-only HEAD~3..HEAD output}

---
Initialized. Ready for user questions.
```

---

## Frontend Details

### UI Components

#### Push-to-Talk Button
- **Size**: Large (200x200px minimum)
- **Behavior**: 
  - Press and hold to record
  - Shows pulsing animation while recording
  - Release to send
  - Disabled during processing
- **States**: Idle, Recording, Processing, Error
- **Safety**: Large enough to use without looking

#### Transcript Display
- Shows user's speech as it's being transcribed
- Editable before sending (future enhancement)
- Scrollable for long questions
- Highlighted border when active

#### Response Display
- Shows cursor's text response
- Auto-scrolls as text appears
- Persistent (doesn't clear between questions)
- Shows conversation history

#### Status Indicator
- Visual + text status
- States: "Ready", "Listening", "Processing", "Speaking", "Error"
- Color-coded for at-a-glance understanding

### Web Speech API Usage

#### Speech Recognition (STT)
```javascript
const recognition = new webkitSpeechRecognition();
recognition.continuous = true;  // Don't stop automatically
recognition.interimResults = true;  // Show partial results
recognition.lang = 'en-US';

// Start on button press
recognition.start();

// Update transcript as user speaks
recognition.onresult = (event) => {
  // Update UI with interim and final results
};

// Stop on button release
recognition.stop();
```

#### Speech Synthesis (TTS)
```javascript
const utterance = new SpeechSynthesisUtterance(responseText);
utterance.rate = 1.0;  // Normal speed
utterance.pitch = 1.0;
utterance.lang = 'en-US';

speechSynthesis.speak(utterance);

// Allow interruption
// User can tap to stop speaking
```

### Mobile Optimizations
- Prevent zoom on button press
- Keep screen awake during session
- Handle phone calls gracefully (pause session)
- Work in background (audio continues)
- Add to home screen capability

---

## Edge Cases & Error Handling

### Backend Errors

#### Cursor Process Crashes
- **Detection**: Monitor process.Wait() in goroutine
- **Action**: Log error, clean up session, return error to frontend
- **User Experience**: "Sorry, lost connection. Please start a new session."

#### Cursor Takes Too Long (>60s)
- **Detection**: Timeout on stdout read
- **Action**: Kill process, restart if needed
- **User Experience**: "That's taking a while, let me try again..."

#### Invalid Session ID
- **Response**: 404 with clear error message
- **User Action**: Frontend auto-restarts session

#### Context Files Missing
- **Behavior**: Graceful degradation, warn in logs
- **User Experience**: Still works, just less context

### Frontend Errors

#### Speech Recognition Not Supported
- **Detection**: Check for webkitSpeechRecognition on load
- **Fallback**: Text input field appears
- **Message**: "Voice not supported, please type"

#### Speech Recognition Fails
- **Common**: Network issues, mic permissions
- **Recovery**: Show error, allow retry
- **Fallback**: Manual text input

#### Network Request Fails
- **Retry**: Automatic retry once after 2 seconds
- **User Action**: Show error, "Retry" button

#### Lost Connection to Server
- **Detection**: Heartbeat fails 3 times
- **Action**: Show reconnection UI
- **Behavior**: Attempt to restore session or create new one

### Session Management Edge Cases

#### User Closes Tab Without Ending Session
- **Detection**: Heartbeat timeout (10 minutes)
- **Action**: Background job cleans up session
- **Summary**: Generate anyway if possible

#### Multiple Tabs Open
- **Behavior**: Each tab gets its own session
- **Limitation**: User should use one tab (document this)

#### Server Restarts
- **Impact**: All active sessions lost
- **Recovery**: Frontend detects on next request, creates new session
- **User Experience**: "Session lost, starting fresh"

---

## Configuration & Environment

### Environment Variables
```bash
CURSOR_API_KEY=sk_xxxxx           # Required: Cursor API key
CODEBASE_PATH=/path/to/code       # Required: Path to codebase
PORT=3000                         # Optional: Server port (default 3000)
SESSION_TIMEOUT_MINUTES=10        # Optional: Inactivity timeout
MAX_CONTEXT_FILES=3               # Optional: Max past conversations to include
```

### Required Files
- `.cursor-voice/system-prompt.md` - Required
- `.cursor-voice/project-overview.md` - Required
- `.cursor-voice/active-context.md` - Optional
- `.cursor-voice/conversation-summaries/` - Created automatically

### Setup Steps
1. Install Cursor CLI: `curl https://cursor.com/install -fsSL | bash`
2. Get Cursor API key from settings
3. Create `.cursor-voice/` folder in codebase
4. Write system-prompt.md and project-overview.md
5. Setup Tailscale on both phone and computer
6. Build and run Go server
7. Access from phone via Tailscale IP

---

## Out of Scope for MVP

These features are intentionally excluded from MVP to ship faster:

### Future Enhancements
- **Edit transcript before sending**: MVP just sends immediately
- **Streaming responses**: MVP waits for full response
- **Multiple concurrent users**: MVP is single-user
- **Persistent storage**: MVP is in-memory only
- **Authentication**: MVP trusts Tailscale security
- **Response caching**: MVP queries fresh each time
- **Voice commands**: No special commands like "repeat that" or "stop"
- **Custom voice selection**: Uses system default
- **Offline mode**: Requires active connection
- **Code file attachments**: Can't send screenshots or files
- **Multi-language support**: English only
- **Analytics/logging**: Minimal logging only

### Known Limitations
- Only works in Chrome/Safari (Web Speech API requirement)
- Requires active internet (cursor-agent needs API access)
- Single codebase per server instance
- No conversation branching or history navigation
- TTS can't be paused/resumed mid-sentence (browser limitation)

---

## Success Metrics

### MVP is successful if:
1. User can start a session in <5 seconds
2. Voice recording captures full rambling (no cutoffs)
3. Cursor correctly identifies context >70% of the time
4. Response latency <30 seconds for typical questions
5. Sessions remain stable for 30+ minute drive
6. Conversation summaries are useful for next session
7. Zero need to look at screen while driving

### Key User Feedback Questions
- Does cursor find the right code context?
- Are responses concise enough for audio?
- Is push-to-talk comfortable while driving?
- Do summaries help continuity across sessions?
- What's missing that would make this essential?

---

## Security Considerations

### Trust Model
- **Tailscale**: Provides encrypted tunnel, handles auth
- **No public exposure**: Server only accessible via Tailscale
- **Code stays local**: Never sent to external services except Cursor API
- **API key security**: Stored in env var, never exposed to frontend

### Risks
- **Phone compromise**: Could access codebase via Tailscale
- **Cursor API**: Code sent to Cursor's service (existing trust)
- **Voice data**: Processed by browser, sent as text only

### Mitigations
- Use strong Tailscale ACLs
- Don't use on public/insecure networks
- Rotate Cursor API key periodically
- Review conversation summaries before committing to git

---

## Development Phases

### Phase 1: Core Backend (Week 1)
- Go server with session management
- cursor-agent process spawning
- Basic API endpoints
- Context file loading

### Phase 2: Basic Frontend (Week 1)
- Next.js app with push-to-talk
- Web Speech API integration
- API client to backend
- Basic UI/styling

### Phase 3: Integration & Testing (Week 2)
- End-to-end testing
- Tailscale setup and testing
- Error handling refinement
- Mobile browser testing

### Phase 4: Polish & Documentation (Week 2)
- Conversation summaries
- Setup documentation
- User guide for creating context files
- Deploy and dogfood

---

## Open Questions

These should be resolved during implementation:

1. **How does cursor-agent chat handle stdin/stdout?**
   - Need to test if responses are line-buffered or need delimiters
   - Confirm it supports long-running interactive sessions

2. **What's the actual format of cursor responses?**
   - Plain text? JSON? Markdown?
   - How to detect end of response?

3. **Can we pass working directory to cursor-agent?**
   - Or does it need to be run from codebase root?

4. **How to handle git repositories with uncommitted changes?**
   - Should initialization mention dirty working tree?

5. **Optimal context file sizes?**
   - How much text before initialization is too verbose?

6. **Rate limits on Cursor API?**
   - Need to handle gracefully if hit limits

---

## Appendix: Example Context Files

### system-prompt.md
```markdown
You are assisting a developer who is DRIVING and cannot see any code or screens.

Critical Rules:
1. Always announce which files/areas you're examining BEFORE giving your answer
2. If the question is ambiguous, list all options you found and ask which one they mean
3. Keep responses conversational and concise - this is audio-only
4. Use natural language for locations ("in your authentication middleware" not "auth.go line 42")
5. Remember context from previous questions in this conversation
6. When context is unclear, prioritize recently modified files
7. Never list long code snippets - summarize and explain instead
8. If you're unsure, say so and ask for clarification

The user will use casual, imprecise language like "that API thing" or "the database stuff". 
Your job is to figure out what they mean and help them think through their code.
```

### project-overview.md
```markdown
# MyApp - E-commerce Platform

## Architecture
- Monolithic Go backend with REST API
- React frontend
- PostgreSQL database
- Redis for caching and sessions

## Key Areas
- **Authentication**: auth/ directory - JWT tokens, session management
- **API Handlers**: api/handlers/ - REST endpoints for products, orders, users
- **Database Layer**: db/ - Models, migrations, queries
- **Background Jobs**: workers/ - Order processing, email sending
- **Payment Integration**: payments/ - Stripe integration

## Current Tech Debt
- Need to refactor auth middleware for better error handling
- Payment webhook verification is fragile
- Missing tests for order processing logic
```

### active-context.md
```markdown
Currently refactoring the payment processing flow to handle edge cases better.

Focus areas:
- payments/webhook.go - Stripe webhook handling
- orders/processor.go - Order state machine
- db/transactions.go - Payment transaction logging

Recent changes:
- Added idempotency keys for Stripe calls
- Fixed race condition in order status updates
```

---

## Document Version
- **Version**: 1.0
- **Date**: October 11, 2024
- **Status**: MVP Specification
- **Next Review**: After Phase 1 completion