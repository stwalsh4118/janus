# V0 Prompt - Janus Voice Interface (Copy-Paste Ready)

---

Create a mobile-first, voice-enabled web interface optimized for safe use while driving. The app is called "Janus - Voice Portal to Your Codebase" and allows developers to ask questions about their codebase using voice input.

## Requirements

### Core Interaction Model

**FULL-SCREEN TOUCH INTERACTION**: The entire page acts as one large press-and-hold button. Users can press anywhere on the screen to start recording. This eliminates the need to aim for a specific button while driving.

### Layout Structure

Three-section layout with full-screen interaction overlay:

1. **Top: Status Bar (fixed, ~60px)**
   - Connection status with colored dot indicator (green=connected, yellow=connecting, red=disconnected)
   - System info: "v0.1.0 • 2 active sessions"
   - Dark background, high contrast text
   - Border bottom separator
   - Z-index above interaction layer

2. **Middle: Scrollable Conversation History**
   - Chat-style message bubbles
   - User messages: right-aligned, primary color background, User avatar icon
   - Assistant messages: left-aligned, secondary/card background, Bot avatar icon
   - Avatars: 40px circular with icons
   - Timestamps: relative time ("2 mins ago"), small text, 70% opacity
   - Play button on assistant messages (Volume icon + "Play")
   - Auto-scroll to latest message
   - Empty state: "Press and hold anywhere to ask a question"
   - Current transcript shows with pulsing animation while recording (semi-transparent, primary border)
   - Scrollable when not interacting, scroll disabled during recording

3. **Bottom: Visual Indicator (fixed, centered)**
   - Large microphone icon (120px) centered at bottom
   - State label below icon
   - Helper text
   - Padding (48px vertical)
   - Z-index above interaction layer

### Full-Screen State Indicator (Primary Element)

**The entire page background changes with an "aura" effect to indicate state:**

**Five Visual States:**

1. **Idle**: 
   - Page background: Default dark (#0a0a0a)
   - Aura: Subtle blue glow around page edges (inset shadow or border gradient)
   - Center icon: MicOff icon (120px), muted color
   - Label: "Press and Hold Anywhere"
   - Helper text: "Touch anywhere on screen to start recording"

2. **Recording**: 
   - Page background: Very subtle red tint (#1a0505)
   - Aura: Pulsing red glow around entire page edge (thick inset shadow, 20-30px, red with 50% opacity)
   - Animation: Pulse ring effect expanding from edges
   - Center icon: Mic icon (120px), red color
   - Label: "Recording..."
   - Helper text: "Speak your question clearly"
   - Page pulses with breathing animation

3. **Processing**: 
   - Page background: Very subtle blue tint (#050a1a)
   - Aura: Steady blue glow around page edge (inset shadow, blue)
   - Center icon: Loader2 icon (120px, rotating), blue color
   - Label: "Processing..."
   - Helper text: "Waiting for response..."
   - Interaction disabled

4. **Speaking**: 
   - Page background: Very subtle green tint (#051a0a)
   - Aura: Pulsing green glow around page edge (inset shadow, green)
   - Center icon: Volume2 icon (120px), green color
   - Label: "Speaking..."
   - Helper text: "Tap anywhere to stop playback"
   - Gentle pulse animation
   - Can tap to interrupt

5. **Error**: 
   - Page background: Subtle red tint (#1a0000)
   - Aura: Red glow around page edge (inset shadow, red)
   - Center icon: AlertCircle icon (120px), red color
   - Label: "Error"
   - Error message in red text below
   - Can tap to dismiss

**Aura Effect Details:**
- Use `box-shadow: inset 0 0 100px 20px rgba(color, opacity)` for glow effect
- Animate opacity for pulsing states (recording, speaking)
- Transition smoothly between states (300ms ease-in-out)
- Can also use pseudo-elements with radial gradients for more dramatic effect

**Interactions:**
- Touch down anywhere: start recording (except on status bar or message play buttons)
- Touch up: stop recording and send
- During speaking: tap anywhere to stop TTS
- During processing: touches ignored
- Prevent text selection and default touch behaviors
- Prevent page scrolling during active recording

### Message Bubbles

**User Messages (right-aligned):**
- Primary color background, white text
- Max width 80%, rounded-2xl, padding 16px, subtle shadow
- 40px circular avatar (User icon, primary background)
- Timestamp at bottom (text-xs, 70% opacity)

**Assistant Messages (left-aligned):**
- Secondary/card background, standard text color
- Max width 80%, rounded-2xl, padding 16px, subtle shadow
- 40px circular avatar (Bot icon, secondary background)
- Timestamp + Play button at bottom (Volume2 icon, interactive)

**Current Transcript (temporary):**
- Similar to user message, primary/20 background, primary border
- Pulsing opacity animation

### Color Scheme (Dark Theme)

```
Background: #0a0a0a (very dark)
Card/Surface: #1a1a1a
Primary: #3b82f6 (blue)
Secondary: #374151 (gray)
Foreground: #fafafa (white)
Muted: #9ca3af (gray)
Destructive: #ef4444 (red)
Success: #22c55e (green)
Border: #27272a
```

**High Contrast:** WCAG AAA compliance (7:1), readable in bright sunlight

### Typography

- Button labels: text-lg font-bold
- Button helper text: text-sm text-muted-foreground
- Messages: text-sm, whitespace-pre-wrap
- Timestamps: text-xs, 70% opacity
- Status bar: text-sm

### Mobile Optimizations

- Responsive: 320px to 768px width
- Portrait and landscape support
- **Touch target: ENTIRE SCREEN** (maximum possible touch area)
- Prevent zoom: viewport user-scalable=no
- Prevent text selection during interaction
- Touch-action: none on main interaction overlay
- Allow scrolling only when in idle state
- Smooth state transitions (300ms)
- Optimize for one-handed use

### Animations

- Page aura: Pulsing inset box-shadow for recording/speaking states
- State transitions: Background and aura change smoothly (300ms ease-in-out)
- Recording: Pulse animation on page edge glow (2s infinite)
- Processing: Smooth rotation of Loader2 icon (center)
- Speaking: Gentle pulse on green aura
- Auto-scroll: Smooth scroll behavior
- Icon transitions: Fade between state icons (200ms)

### Tech Stack

- Next.js 14 with App Router
- TypeScript with "use client" directive
- Tailwind CSS
- shadcn/ui components: Button, Card, ScrollArea
- Lucide React icons: Mic, MicOff, Loader2, Volume2, User, Bot
- Must work on iOS Safari and Android Chrome

### Layout Reference

```
┌─────────────────────────────────────┐
│ ● Connected  v0.1.0 • 2 sessions   │ ← Fixed Header (clickable)
├─────────────────────────────────────┤
│ ╔═══════════════════════════════╗  │
│ ║                               ║  │ ← Glowing aura (state-based)
│ ║  [Bot] Assistant message      ║  │   changes color/pulse
│ ║        "Response here..."     ║  │
│ ║        2 mins ago [🔊 Play]  ║  │
│ ║                               ║  │
│ ║       User question [User]    ║  │ ← Scrollable Messages
│ ║       "How does X work?"      ║  │   (entire area is interactive)
│ ║       5 mins ago              ║  │
│ ║                               ║  │
│ ║  [Bot] Another response...    ║  │
│ ║        Just now [🔊 Play]    ║  │
│ ║                               ║  │
│ ║                               ║  │
│ ║                               ║  │
│ ║          [🎤 120px]           ║  │ ← Large icon indicator
│ ║      Press and Hold           ║  │   (not a button, visual only)
│ ║         Anywhere              ║  │
│ ║                               ║  │
│ ╚═══════════════════════════════╝  │
└─────────────────────────────────────┘
     ↑
     Entire shaded area is one large touch target
     Aura glow indicates current state
```

**State Visualization:**
- **Idle**: Blue aura glow (subtle)
- **Recording**: Red aura with pulse animation (prominent)
- **Processing**: Blue aura (steady)
- **Speaking**: Green aura with pulse (gentle)
- **Error**: Red aura (steady)

### Key Design Goals

1. **Full-screen interaction** - entire page is one large touch target (ultimate driving safety)
2. **Aura-based state feedback** - glowing page edges indicate current state
3. **No aiming required** - press anywhere to interact
4. **Clear visual states** - page background and edge glow change with state
5. **High contrast** for outdoor visibility
6. **Professional, modern design**
7. **Fast, smooth animations** (60fps)
8. **Accessible** with proper ARIA labels

### Technical Implementation Notes

**Full-Screen Touch Overlay:**
- Create a full-viewport overlay div that captures touch events
- Use fixed positioning with inset-0
- Layer below status bar and message play buttons (z-index management)
- Apply `touch-action: none` to prevent default behaviors
- Disable scrolling during recording state
- Apply state-based styling (background tint + inset box-shadow for aura)

**Aura Effect Implementation:**
```css
/* Example for recording state */
.page-recording {
  background: linear-gradient(to bottom, #0a0a0a, #1a0505);
  box-shadow: inset 0 0 100px 30px rgba(239, 68, 68, 0.5);
  animation: pulse-aura 2s ease-in-out infinite;
}

@keyframes pulse-aura {
  0%, 100% { box-shadow: inset 0 0 100px 30px rgba(239, 68, 68, 0.3); }
  50% { box-shadow: inset 0 0 120px 40px rgba(239, 68, 68, 0.6); }
}
```

### Components to Generate

1. Main page layout with full-screen interaction overlay
2. FullScreenInteraction component with state-based aura effects
3. StateIndicator component (centered icon + labels, not interactive)
4. MessageBubble component (user and assistant variants)
5. ConversationHistory component with auto-scroll
6. StatusIndicator component

Please use TypeScript, proper type safety, shadcn/ui components, and Tailwind utility classes. Ensure the full-screen touch interaction works smoothly on both iOS Safari and Android Chrome.

