# PBI-4: Voice-Enabled Frontend Interface

[View in Backlog](../backlog.md#user-content-4)

## Overview

Build a Next.js web application with voice input/output capabilities using Web Speech API. The interface provides a hands-free, audio-only experience optimized for use while driving, featuring push-to-talk recording, real-time transcription, and text-to-speech responses.

## Problem Statement

Developers need to:
- Ask questions about their codebase while driving without looking at screens
- Record long, rambling questions without time limits or cutoffs
- Hear responses read aloud automatically
- Interact with large touch targets that don't require visual precision
- Access the system from their phone via Tailscale connection

## User Stories

- As a developer, I want to hold a button to record my question so that I can speak naturally without typing
- As a developer, I want to see my speech transcribed in real-time so that I can verify it's capturing correctly (quick glance)
- As a developer, I want responses read aloud automatically so that I never need to read text while driving
- As a developer, I want large touch targets so that I can safely interact without looking at the screen
- As a developer, I want the app to work on my phone's browser so that I don't need to install anything

## Technical Approach

### Technology Stack

- **Framework**: Next.js 14 with App Router
- **Styling**: Tailwind CSS
- **Voice Input**: Web Speech API (webkitSpeechRecognition)
- **Voice Output**: Web Speech API (speechSynthesis)
- **HTTP Client**: fetch API for backend communication
- **Deployment**: Static export served by Go backend

### Core Components

#### 1. Push-to-Talk Button
```typescript
// Large (200x200px), press-and-hold interaction
// States: idle, recording, processing, speaking, error
// Visual feedback: pulsing animation while recording
// Haptic feedback on press/release (if available)
```

#### 2. Speech Recognition Integration
```javascript
const recognition = new webkitSpeechRecognition();
recognition.continuous = true;        // No automatic cutoff
recognition.interimResults = true;    // Show partial results
recognition.lang = 'en-US';

// Start on button press, stop on release
// Handle errors and fallback to text input
```

#### 3. Speech Synthesis Integration
```javascript
const utterance = new SpeechSynthesisUtterance(responseText);
utterance.rate = 1.0;
utterance.pitch = 1.0;
utterance.lang = 'en-US';

// Allow interruption via tap
// Queue management for multiple responses
```

#### 4. Session Manager
- Creates session on app load
- Stores session_id in component state
- Sends heartbeat every 30 seconds
- Handles session expiration and reconnection

#### 5. API Client
```typescript
class CursorVoiceClient {
  async startSession(): Promise<string>;
  async ask(sessionId: string, question: string): Promise<string>;
  async endSession(sessionId: string): Promise<void>;
  async heartbeat(sessionId: string): Promise<void>;
}
```

### UI Layout

```
┌─────────────────────────────┐
│     Status: Ready           │ ← Status indicator
├─────────────────────────────┤
│                             │
│                             │
│     ┌─────────────┐        │
│     │             │        │
│     │   [HOLD]    │        │ ← Large push-to-talk button
│     │             │        │
│     └─────────────┘        │
│                             │
├─────────────────────────────┤
│ "How does the auth          │ ← Transcript display
│  middleware work?"          │
├─────────────────────────────┤
│ "I'm looking at your        │ ← Response display
│  authentication middleware  │   (scrollable)
│  in middleware/jwt.go..."   │
└─────────────────────────────┘
```

### Mobile Optimizations

- Prevent zoom on button press (viewport meta tag)
- Keep screen awake during active session (Wake Lock API)
- Handle phone calls gracefully (pause on visibility change)
- Add to home screen capability (PWA manifest)
- Large touch targets (min 44x44pt)
- High contrast for outdoor visibility

### Error Handling & Fallbacks

1. **Speech Recognition Not Supported**
   - Show text input field as fallback
   - Display warning message

2. **Microphone Permission Denied**
   - Show permission instructions
   - Fallback to text input

3. **Network Errors**
   - Auto-retry once after 2 seconds
   - Show error message with retry button
   - Maintain session state during retry

4. **Session Lost**
   - Detect via heartbeat failure (3 consecutive)
   - Automatically create new session
   - Show "Reconnecting..." status

## UX/UI Considerations

### Safety First
- Large touch targets that don't require looking
- Audio feedback for all interactions
- Minimize visual distraction
- Support voice-only workflow entirely

### Conversational Flow
- Show conversation history
- Clear visual states (listening, processing, speaking)
- Allow interruption of speech output
- Preserve context across questions

### Mobile-Optimized
- Fast load time (<2s)
- Responsive layout (320px to 768px)
- Works in both portrait and landscape
- Handles screen rotation gracefully

## Acceptance Criteria

- [ ] Next.js app builds and runs successfully
- [ ] App detects Web Speech API support on load
- [ ] Push-to-talk button starts recording on press
- [ ] Speech recognition captures audio continuously while held
- [ ] Transcript updates in real-time as user speaks
- [ ] Button release stops recording and sends question
- [ ] API client creates session on app load
- [ ] Question is sent to backend with session_id
- [ ] Response is displayed in text area
- [ ] Response is automatically spoken via TTS
- [ ] User can tap to stop TTS mid-sentence
- [ ] Heartbeat is sent every 30 seconds
- [ ] Session is recreated if backend connection fails
- [ ] Fallback text input shown if speech not supported
- [ ] Button is large enough for eyes-free interaction (200x200px)
- [ ] App works on iOS Safari and Android Chrome
- [ ] Visual states clearly indicate system status
- [ ] Conversation history is preserved during session
- [ ] App prevents zoom on button interactions
- [ ] Loading states show clear feedback
- [ ] Error messages are clear and actionable

## Dependencies

- PBI-1 (Backend Session Management) completed
- Node.js and pnpm installed for development
- Mobile device with Tailscale for testing
- Browser with Web Speech API support (iOS Safari, Chrome)

## Open Questions

- Should we implement wake lock to keep screen on?
- Do we need offline capability or always require connection?
- Should transcript be editable before sending?
- How do we handle very long responses (>5 minutes TTS)?
- Should we implement voice commands like "stop" or "repeat"?
- Do we need user authentication or rely on Tailscale?

## Related Tasks

Tasks will be created when PBI moves to "Agreed" status.

