# Design Spec & v0 Prompt for Janus Voice Interface

## Document Purpose
This document contains a complete design specification and prompt for v0.dev to generate the styling and visual design for the Janus voice-enabled frontend interface.

---

## V0 PROMPT

Create a mobile-first, voice-enabled web interface optimized for safe use while driving. The interface should be called "Janus - Voice Portal to Your Codebase" and allows developers to ask questions about their codebase using voice input.

### Core Interaction Model

**FULL-SCREEN TOUCH INTERACTION**: The entire page acts as one large press-and-hold button. Users can press anywhere on the screen to start recording. This eliminates the need to aim for a specific button while driving, providing the ultimate safety feature.

### Core Layout

The interface has three main sections with a full-screen interaction overlay:

1. **Top: Status Bar** (fixed header)
   - Shows connection status with colored indicator
   - Displays system health info (version, active sessions)
   - Compact, minimal height (~60px)
   - Dark background with high contrast text

2. **Middle: Conversation History** (scrollable area)
   - Chat-style message bubbles
   - User messages aligned right with primary color background
   - Assistant responses aligned left with secondary/muted background
   - Avatar icons (User icon for user, Bot icon for assistant)
   - Timestamps showing relative time ("2 mins ago")
   - Empty state: "No messages yet / Hold the button to ask a question"
   - Current transcript shows in real-time with pulsing animation while recording
   - Auto-scrolls to latest message
   - Messages have subtle shadows and rounded corners

3. **Bottom: Visual State Indicator** (fixed footer, centered)
   - Large icon (120px) indicating current state
   - State label and helper text
   - NOT a button - purely visual feedback
   - Padding: 48px vertical

### Full-Screen Interaction Design

This is the primary interaction element - the entire page surface responds to touch:

**Interaction Model:**
- ENTIRE page is one large touch target
- Press and hold anywhere to start recording
- Release to stop and send
- No need to aim for specific button
- Maximum safety for driving

**Page-Level "Aura" Visual States:**

The entire page changes appearance with a glowing "aura" effect around the edges to indicate current state:

1. **Idle State** (default):
   - Page background: Default dark (#0a0a0a)
   - Aura: Subtle blue glow around page edges (inset box-shadow)
   - Center icon: MicOff icon (120px), muted color
   - Label: "Press and Hold Anywhere"
   - Helper text: "Touch anywhere on screen to start recording"

2. **Recording State** (actively recording):
   - Page background: Very subtle red tint (#1a0505)
   - Aura: Pulsing red glow around entire page edge (thick inset shadow, 20-30px, red 50% opacity)
   - Animation: Pulse ring effect expanding from edges (breathing effect)
   - Center icon: Mic icon (120px), red color
   - Label: "Recording..."
   - Helper text: "Speak your question clearly"
   - Entire page pulses with breathing animation

3. **Processing State** (waiting for response):
   - Page background: Very subtle blue tint (#050a1a)
   - Aura: Steady blue glow around page edge (inset shadow, blue)
   - Center icon: Loader2 icon (120px, rotating animation), blue color
   - Label: "Processing..."
   - Helper text: "Waiting for response..."
   - Touch interaction disabled

4. **Speaking State** (TTS active):
   - Page background: Very subtle green tint (#051a0a)
   - Aura: Pulsing green glow around page edge (inset shadow, green)
   - Center icon: Volume2 icon (120px), green color
   - Label: "Speaking..."
   - Helper text: "Tap anywhere to stop playback"
   - Gentle pulse animation
   - Can tap to interrupt playback

5. **Error State**:
   - Page background: Subtle red tint (#1a0000)
   - Aura: Red glow around page edge (inset shadow, red)
   - Center icon: AlertCircle icon (120px), red color
   - Label: "Error"
   - Error message in red text below
   - Can tap to dismiss

**Aura Effect Technical Details:**
- Use `box-shadow: inset 0 0 100px 20px rgba(color, opacity)` for glow effect
- Animate opacity for pulsing states (recording, speaking) using keyframe animation
- Transition smoothly between states (300ms ease-in-out)
- Can also use pseudo-elements (::before, ::after) with radial gradients for more dramatic effect
- Consider using CSS filters (blur, brightness) for enhanced glow

**Interaction Feedback:**
- Touch down anywhere: start recording (except status bar and play buttons)
- Touch up: stop recording and send question
- During speaking: tap anywhere to stop TTS
- During processing: all touches ignored
- Haptic feedback on supported devices (not visually indicated)
- Prevent text selection and default touch behaviors
- Prevent page scrolling during active recording

### Message Bubble Design

**User Messages (right-aligned):**
- Background: Primary color (matching button idle state)
- Text: Primary foreground color (white)
- Max width: 80% of container
- Padding: 16px
- Border radius: Large (rounded-lg or rounded-2xl)
- Shadow: Subtle
- Avatar: Circular 40px with User icon, primary color background
- Timestamp: Small text, 70% opacity, bottom of bubble

**Assistant Messages (left-aligned):**
- Background: Secondary/muted color (gray-800 or card background)
- Text: Standard foreground color
- Max width: 80% of container
- Padding: 16px
- Border radius: Large (rounded-lg or rounded-2xl)
- Shadow: Subtle
- Avatar: Circular 40px with Bot icon, secondary color background
- Timestamp + Play button: Small text with volume icon, bottom of bubble
- Play button: "Play" with volume icon, interactive, hover effect

**Current Transcript (temporary, while recording):**
- Similar to user message but with semi-transparent background (primary/20)
- Border: Primary color border
- Pulsing opacity animation to indicate it's temporary/in-progress

### Color Scheme

**Dark Theme** (primary theme for driving, reduce eye strain):
- Background: Very dark (black or near-black, #000000 or #0a0a0a)
- Card/Surface: Dark gray (#1a1a1a or #1f1f1f)
- Primary: Blue-purple (#3b82f6 or #6366f1) - for buttons and user messages
- Secondary: Muted gray (#374151 or #4b5563)
- Accent: Use for highlighting
- Text: High contrast white/near-white
- Muted text: Gray (#9ca3af)
- Destructive: Red (#ef4444)
- Success/Speaking: Green (#22c55e)
- Recording: Red (#ef4444)

**High Contrast:**
- All text must be easily readable in bright sunlight
- Minimum contrast ratio: WCAG AAA (7:1 for normal text)
- Status indicators should be color + icon for accessibility

### Status Indicator Design

Located at top of screen:

- Compact horizontal bar with padding
- Connection status with colored dot:
  - Green dot: "Connected"
  - Yellow dot: "Connecting..."
  - Red dot: "Disconnected"
  - Gray dot: "Error"
- Additional info in smaller text: "v0.1.0 • 2 active sessions"
- Border bottom to separate from content

### Typography

- Font family: System font stack (san-serif, optimized for mobile)
- Button labels: Large, bold (text-lg font-bold)
- Button helper text: Small, muted (text-sm text-muted-foreground)
- Message content: Base size (text-sm), readable, whitespace preserved
- Timestamps: Extra small (text-xs), 70% opacity
- Status bar: Small to medium (text-sm)

### Mobile Optimizations

**Responsive:**
- Works 320px to 768px width
- Portrait and landscape modes
- Touch targets: Minimum 44x44pt (Apple HIG)
- Button maintains 200x200px in both orientations

**Performance:**
- Fast initial load
- Smooth animations (60fps)
- Efficient scrolling with many messages
- Prevent zoom on double-tap (viewport: user-scalable=no)

**Accessibility:**
- Prevent text selection on button
- Touch action: none on button to prevent conflicts
- Webkit touch callout: none
- High contrast mode compatible
- Screen reader friendly (proper ARIA labels)

### Animations

- Button press: Scale transform (95%) with smooth transition (200ms)
- Recording pulse: Continuous pulse animation with ring glow
- Spinner: Smooth rotation
- Auto-scroll: Smooth scroll behavior
- Message appearance: Subtle fade-in
- State transitions: 200ms ease-in-out

### Empty States

**No Messages:**
- Center-aligned text
- Muted color
- Icon optional (microphone or conversation icon)
- Text: "No messages yet" / "Hold the button to ask a question"

**No Connection:**
- Status bar shows red/disconnected
- Button disabled with explanation
- Helper text: "Connect to backend to start"

### Additional UI Elements

**Loading States:**
- Skeleton loaders for messages if needed
- Spinner in button during processing
- Progress indication if applicable

**Error Displays:**
- Inline errors below button (red text)
- Toast/alert for critical errors (optional)
- Clear, actionable error messages

### Technical Constraints

- Use Tailwind CSS for all styling
- Built with Next.js 14 + React
- Use shadcn/ui components where applicable:
  - Button component for main button
  - Card component for message bubbles
  - ScrollArea component for conversation history
- Lucide React icons:
  - Mic, MicOff for button states
  - Loader2 for loading spinner
  - Volume2 for play audio
  - User for user avatar
  - Bot for assistant avatar
- Must work on iOS Safari and Android Chrome
- Should work without JavaScript for basic layout (progressive enhancement)

### Reference Layout (ASCII)

```
┌─────────────────────────────────────┐
│ ● Connected  v0.1.0 • 2 sessions   │ ← Status Bar (60px, clickable)
├─────────────────────────────────────┤
│ ╔═══════════════════════════════╗  │
│ ║                               ║  │ ← Glowing aura (inset shadow)
│ ║  [Bot] Assistant message      ║  │   Changes color by state
│ ║        "The auth middleware..." ║│   Red = recording
│ ║        2 mins ago [🔊 Play]  ║  │   Green = speaking
│ ║                               ║  │   Blue = idle/processing
│ ║       User message [User]     ║  │
│ ║       "How does auth work?"   ║  │ ← Scrollable Messages
│ ║       5 mins ago              ║  │   (entire area is interactive)
│ ║                               ║  │
│ ║  [Bot] Another response...    ║  │
│ ║        Just now [🔊 Play]    ║  │
│ ║                               ║  │
│ ║       [Recording...] [User]   ║  │ ← Current transcript
│ ║       (pulsing)               ║  │   (while speaking)
│ ║                               ║  │
│ ║                               ║  │
│ ║          [🎤 120px]           ║  │ ← Visual indicator
│ ║    Press and Hold Anywhere    ║  │   (NOT a button)
│ ║                               ║  │
│ ╚═══════════════════════════════╝  │
└─────────────────────────────────────┘
     ↑
     Entire shaded area (╔═══╗) is ONE large touch target
     Aura glow pulses during recording/speaking
     Background tint changes subtly with state
```

### Design Goals Summary

1. **Safety First**: ENTIRE screen is touch target - no aiming required, minimal visual attention needed
2. **Aura-Based Feedback**: Page-level visual states with glowing edges for peripheral awareness
3. **Clear Feedback**: Visual states for every interaction through background and edge glow
4. **Mobile Optimized**: Fast, responsive, works in all orientations
5. **High Contrast**: Readable in bright sunlight while driving
6. **Conversational**: Natural chat-like interface
7. **Accessible**: Works for users with various abilities
8. **Professional**: Clean, modern design that developers will trust
9. **No Aiming Required**: Press anywhere - perfect for eyes-free interaction

### Use Cases to Design For

1. Developer driving, needs to ask question without looking
2. Quick glance to verify transcript is correct
3. Listening to long response while keeping eyes on road
4. Stopping TTS mid-sentence if needed
5. Reviewing conversation history at red light
6. Handling connection issues gracefully

---

## IMPLEMENTATION NOTES

When generating code with v0:
- Use TypeScript
- Use "use client" directive for interactive components
- Follow Next.js 14 App Router patterns
- Use Tailwind utility classes
- Import shadcn/ui components correctly
- Ensure proper type safety
- Include proper ARIA labels
- Test on mobile viewport sizes

### Full-Screen Touch Overlay Implementation

**Structure:**
```typescript
<div className="relative min-h-screen">
  {/* Full-screen interaction overlay with state-based styling */}
  <div 
    className={`fixed inset-0 transition-all duration-300 ${getStateClasses()}`}
    style={{ touchAction: 'none' }}
    onTouchStart={handleTouchStart}
    onTouchEnd={handleTouchEnd}
    onMouseDown={handleMouseDown}
    onMouseUp={handleMouseUp}
  >
    {/* Status bar (with higher z-index, NOT affected by overlay) */}
    <div className="relative z-50">
      <StatusIndicator />
    </div>
    
    {/* Conversation history */}
    <div className="relative z-10">
      <ConversationHistory />
    </div>
    
    {/* Visual state indicator (centered at bottom) */}
    <div className="relative z-10 fixed bottom-0 left-0 right-0">
      <StateIndicator state={currentState} />
    </div>
  </div>
</div>
```

**State Class Generation:**
```typescript
const getStateClasses = () => {
  switch (state) {
    case 'idle':
      return 'bg-[#0a0a0a] shadow-[inset_0_0_100px_20px_rgba(59,130,246,0.3)]';
    case 'recording':
      return 'bg-[#1a0505] shadow-[inset_0_0_100px_30px_rgba(239,68,68,0.5)] animate-pulse-aura';
    case 'processing':
      return 'bg-[#050a1a] shadow-[inset_0_0_100px_20px_rgba(59,130,246,0.4)]';
    case 'speaking':
      return 'bg-[#051a0a] shadow-[inset_0_0_100px_20px_rgba(34,197,94,0.4)] animate-pulse-gentle';
    case 'error':
      return 'bg-[#1a0000] shadow-[inset_0_0_100px_20px_rgba(239,68,68,0.5)]';
  }
};
```

**Z-Index Management:**
- Full-screen overlay: z-0 (base layer)
- Conversation history: z-10 (above overlay, scrollable)
- State indicator: z-10 (above overlay, fixed)
- Status bar: z-50 (top layer, always clickable)
- Message play buttons: z-20 (clickable, above conversation)

**Touch Event Handling:**
- Prevent default on touch events to avoid scrolling during recording
- Allow touch events on specific interactive elements (status bar, play buttons)
- Stop propagation on exempt elements so they don't trigger recording

## STYLE TOKENS REFERENCE

If creating a theme configuration:

```typescript
colors: {
  background: "#0a0a0a",
  foreground: "#fafafa",
  primary: {
    DEFAULT: "#3b82f6",
    foreground: "#ffffff",
  },
  secondary: {
    DEFAULT: "#374151",
    foreground: "#fafafa",
  },
  destructive: {
    DEFAULT: "#ef4444",
    foreground: "#ffffff",
  },
  muted: {
    DEFAULT: "#1f1f1f",
    foreground: "#9ca3af",
  },
  card: {
    DEFAULT: "#1a1a1a",
    foreground: "#fafafa",
  },
  border: "#27272a",
}

spacing: {
  "button-size": "200px",
}

animation: {
  "pulse-aura": "pulse-aura 2s ease-in-out infinite",
  "pulse-gentle": "pulse-gentle 3s ease-in-out infinite",
}

keyframes: {
  "pulse-aura": {
    "0%, 100%": { 
      boxShadow: "inset 0 0 100px 30px rgba(239, 68, 68, 0.3)",
    },
    "50%": { 
      boxShadow: "inset 0 0 120px 40px rgba(239, 68, 68, 0.6)",
    },
  },
  "pulse-gentle": {
    "0%, 100%": { 
      boxShadow: "inset 0 0 100px 20px rgba(34, 197, 94, 0.3)",
    },
    "50%": { 
      boxShadow: "inset 0 0 110px 25px rgba(34, 197, 94, 0.5)",
    },
  },
}
```

---

## DELIVERABLES EXPECTED FROM V0

1. Complete styled component code for:
   - Main page layout with full-screen interaction overlay
   - FullScreenInteraction component with state-based aura effects
   - StateIndicator component (centered visual indicator, not interactive)
   - Message bubbles (user and assistant)
   - ConversationHistory component with auto-scroll
   - StatusIndicator component

2. Tailwind configuration updates:
   - Custom animations for aura pulse effects
   - Custom keyframes for box-shadow animations
   - Any additional color tokens

3. Custom CSS animations:
   - `pulse-aura` for recording state (red glow)
   - `pulse-gentle` for speaking state (green glow)
   - Smooth state transitions

4. Mobile-responsive styles:
   - Full-screen touch overlay
   - Z-index management for layering
   - Touch event handling
   - Prevent default behaviors

5. Dark theme implementation with aura effects

---

## VALIDATION CHECKLIST

After receiving code from v0, verify:

- [ ] Entire page responds to touch (full-screen interaction)
- [ ] All 5 visual states show distinct aura effects
- [ ] Aura glow appears around page edges (inset box-shadow)
- [ ] Recording state has pulsing red aura
- [ ] Speaking state has pulsing green aura
- [ ] Processing state has steady blue aura
- [ ] Page background has subtle color tint per state
- [ ] State transitions are smooth (300ms)
- [ ] Center icon changes with state (120px size)
- [ ] Message bubbles are properly styled
- [ ] Layout works on mobile (320px-768px)
- [ ] High contrast in dark theme
- [ ] Animations are smooth (60fps)
- [ ] Touch works anywhere on screen (maximum touch target)
- [ ] Text is readable in all states
- [ ] Empty states are handled
- [ ] Error states are clear with red aura
- [ ] Loading states are present
- [ ] Timestamp formatting works
- [ ] Auto-scroll behavior works
- [ ] Status bar remains clickable (z-index correct)
- [ ] Play buttons work (don't trigger recording)
- [ ] Conversation history scrolls correctly
- [ ] Scrolling disabled during recording
- [ ] Touch events prevented during processing

