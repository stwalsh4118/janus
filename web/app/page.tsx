"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { StatusBar } from "@/components/status-bar";
import { ConversationHistory } from "@/components/conversation-history";
import { StateIndicator } from "@/components/state-indicator";
import { DebugPanel } from "@/components/debug-panel";
import { apiClient } from "@/lib/api-client";
import { useAudioRecorder } from "@/hooks/useAudioRecorder";
import { useSpeechSynthesis } from "@/hooks/useSpeechSynthesis";
import { generateUUID } from "@/lib/uuid";
import type { AppState, Message, HealthResponse } from "@/lib/types";

export default function Home() {
  // Load eruda mobile console for debugging on phone
  useEffect(() => {
    if (typeof window !== 'undefined') {
      import('eruda').then(eruda => {
        eruda.default.init();
        console.log('[Eruda] Mobile console initialized - tap the console icon to view logs');
      });
    }
  }, []);
  const [state, setState] = useState<AppState>("idle");
  const [messages, setMessages] = useState<Message[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [healthData, setHealthData] = useState<HealthResponse | null>(null);
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [isDebugOpen, setIsDebugOpen] = useState(false);
  const [debugText, setDebugText] = useState<string>("");

  const containerRef = useRef<HTMLDivElement>(null);

  // Audio recorder hook (replaces speech recognition)
  const {
    isSupported,
    isRecording,
    isTranscribing,
    error: recordingError,
    startRecording: startAudioRecording,
    stopRecording: stopAudioRecording,
    resetRecording,
  } = useAudioRecorder();

  // Speech synthesis hook
  const {
    isSupported: ttsSupported,
    isSpeaking,
    isGenerating: isGeneratingAudio,
    currentProvider,
    preferredProvider,
    setPreferredProvider,
    speak,
    stop: stopSpeaking,
    unlockAudio,
  } = useSpeechSynthesis();

  // No transcript display during recording (server-side transcription)

  // Check backend health on mount
  useEffect(() => {
    const checkHealth = async () => {
      try {
        setIsConnected(false);
        const data = await apiClient.healthCheck();
        setHealthData(data);
        setIsConnected(true);
        setError(null);
      } catch (err) {
        console.error("Health check failed:", err);
        setIsConnected(false);
        setError(err instanceof Error ? err.message : "Failed to connect to backend");
      }
    };

    checkHealth();
    // Re-check every 30 seconds
    const interval = setInterval(checkHealth, 30000);
    return () => clearInterval(interval);
  }, []);

  // Create session when connected
  useEffect(() => {
    if (isConnected && !sessionId) {
      const createSession = async () => {
        try {
          const id = await apiClient.startSession();
          setSessionId(id);
        } catch (err) {
          console.error("Failed to create session:", err);
          setError(err instanceof Error ? err.message : "Failed to create session");
        }
      };
      createSession();
    }
  }, [isConnected, sessionId]);

  // Send heartbeat every 30 seconds
  useEffect(() => {
    if (!sessionId) return;

    const heartbeat = async () => {
      try {
        await apiClient.heartbeat(sessionId);
      } catch (err) {
        console.error("Heartbeat failed:", err);
      }
    };

    const interval = setInterval(heartbeat, 30000);
    return () => clearInterval(interval);
  }, [sessionId]);

  // Cleanup session on unmount or page unload
  useEffect(() => {
    if (!sessionId) return;

    const endSessionCleanup = async () => {
      try {
        await apiClient.endSession(sessionId);
      } catch (err) {
        console.error("Failed to end session:", err);
      }
    };

    // Handle page unload (close tab, navigate away, refresh)
    const handleBeforeUnload = () => {
      // Use sendBeacon for reliable cleanup on page unload
      const url = `/api/session/end?session_id=${encodeURIComponent(sessionId)}`;

      if (navigator.sendBeacon) {
        navigator.sendBeacon(url);
      } else {
        // Fallback for browsers that don't support sendBeacon
        fetch(url, { method: "POST", keepalive: true }).catch(() => {
          // Ignore errors during unload
        });
      }
    };

    window.addEventListener("beforeunload", handleBeforeUnload);

    // Cleanup when component unmounts (e.g., navigation in SPA)
    return () => {
      window.removeEventListener("beforeunload", handleBeforeUnload);
      endSessionCleanup();
    };
  }, [sessionId]);

  // Update error state based on recording error - but don't block the entire UI
  useEffect(() => {
    if (recordingError) {
      console.warn('Audio recording error:', recordingError);
      // Don't set the global error state - just log it
      // The user can still use the debug panel for text input
    }
  }, [recordingError]);

  // Update state based on audio recording, transcription, and TTS status
  useEffect(() => {
    if (isSpeaking || isGeneratingAudio) {
      setState("speaking");
    } else if (isRecording) {
      setState("recording");
    } else if (isTranscribing) {
      setState("processing");
    } else if (state === "speaking" || state === "recording" || state === "processing") {
      // Only reset to idle if we were in an active state
      setState("idle");
    }
  }, [isSpeaking, isGeneratingAudio, isRecording, isTranscribing, state]);

  const processQuestion = useCallback(
    async (question: string) => {
      if (!sessionId || !question) return;

      setState("processing");

      try {
        setError(null);

        // Add user message
        const userMessage: Message = {
          id: generateUUID(),
          role: "user",
          content: question,
          timestamp: new Date(),
        };
        setMessages((prev) => [...prev, userMessage]);

        // Get answer
        const answer = await apiClient.ask(sessionId, question);

        // Add assistant message
        const assistantMessage: Message = {
          id: generateUUID(),
          role: "assistant",
          content: answer,
          timestamp: new Date(),
        };
        setMessages((prev) => [...prev, assistantMessage]);

        // Auto-play response with TTS
        if (ttsSupported && answer) {
          console.log('[TTS] Attempting to speak response...');
          try {
            await speak(answer);
            console.log('[TTS] Speak completed');
          } catch (speakErr) {
            console.warn('[TTS] Autoplay blocked (expected on mobile):', speakErr);
            // Set to idle so user can interact
            setState("idle");
          }
        } else {
          console.log('[TTS] TTS not supported or no answer, setting idle');
          setState("idle");
        }
      } catch (err) {
        console.error("Failed to ask question:", err);
        setError(err instanceof Error ? err.message : "Failed to get response");
        setState("error");
      }
    },
    [sessionId, ttsSupported, speak]
  );

  const handlePlayMessage = useCallback(
    async (content: string) => {
      // If already speaking, stop and don't restart
      // This prevents race conditions from rapid clicks
      if (state === "speaking" || isSpeaking) {
        console.log('[PlayMessage] Already speaking, stopping instead');
        stopSpeaking();
        return;
      }
      
      console.log('[PlayMessage] Starting playback');
      await speak(content);
    },
    [state, isSpeaking, speak, stopSpeaking]
  );

  const startRecording = useCallback(() => {
    if (!isSupported) {
      console.warn("Audio recording is not supported in this browser/context");
      // Don't enter error state - user can still use debug panel
      return;
    }

    // Unlock audio on first user interaction (for Safari autoplay)
    unlockAudio();

    resetRecording();
    startAudioRecording();
    setState("recording");
    setError(null);
  }, [isSupported, resetRecording, startAudioRecording, unlockAudio]);

  const stopRecording = useCallback(async () => {
    try {
      setState("processing");
      
      // Stop recording and get transcription from server
      const transcribedText = await stopAudioRecording();
      
      if (!transcribedText) {
        console.warn("No transcription received");
        setState("idle");
        return;
      }

      // Process the transcribed question
      await processQuestion(transcribedText);
    } catch (err) {
      console.error("Failed to process recording:", err);
      setError(err instanceof Error ? err.message : "Failed to process recording");
      setState("error");
    }
  }, [stopAudioRecording, processQuestion]);

  const handleDebugSubmit = useCallback(async () => {
    if (!debugText.trim() || !sessionId) return;

    // Unlock audio on first user interaction (for Safari autoplay)
    unlockAudio();

    const question = debugText.trim();
    setDebugText("");

    await processQuestion(question);
  }, [debugText, sessionId, processQuestion, unlockAudio]);

  // Full-screen interaction handlers
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const handleInteractionClick = (e: MouseEvent | TouchEvent) => {
      const target = e.target as HTMLElement;
      // Skip if clicking on interactive elements
      if (target.closest("[data-interactive]")) {
        return;
      }

      if (state === "idle") {
        // Only try to start recording if speech is supported
        if (isSupported) {
          startRecording();
        }
      } else if (state === "recording") {
        stopRecording();
      } else if (state === "speaking") {
        stopSpeaking();
        setState("idle");
      } else if (state === "error") {
        setError(null);
        setState("idle");
      }
    };

    container.addEventListener("click", handleInteractionClick);

    return () => {
      container.removeEventListener("click", handleInteractionClick);
    };
  }, [state, startRecording, stopRecording, stopSpeaking]);

  const getAuraClasses = () => {
    switch (state) {
      case "recording":
        return "bg-recording shadow-aura-recording animate-pulse-aura";
      case "processing":
        return "bg-processing shadow-aura-processing";
      case "speaking":
        return "bg-speaking shadow-aura-speaking animate-pulse-aura-gentle";
      case "error":
        return "bg-error shadow-aura-error";
      case "idle":
      default:
        return "bg-idle shadow-aura-idle";
    }
  };

  return (
    <div
      ref={containerRef}
      className={`relative h-screen w-full overflow-hidden transition-all duration-300 ease-in-out ${getAuraClasses()}`}
      style={{
        userSelect: "none",
        WebkitUserSelect: "none",
      }}
    >
      <div className="relative z-10 flex h-full min-h-0 flex-col">
        <StatusBar
          isConnected={isConnected}
          activeSessions={healthData?.active_sessions || 0}
          version={healthData?.version}
          ttsProvider={preferredProvider}
          onTtsProviderChange={setPreferredProvider}
          isDebugOpen={isDebugOpen}
          onDebugToggle={() => setIsDebugOpen(!isDebugOpen)}
        />

        <DebugPanel
          isOpen={isDebugOpen}
          sessionId={sessionId}
          debugText={debugText}
          onDebugTextChange={setDebugText}
          onDebugSubmit={handleDebugSubmit}
          isSending={state === "processing"}
          ttsProvider={preferredProvider}
          currentProvider={currentProvider}
          onTtsProviderChange={setPreferredProvider}
        />

        <ConversationHistory
          messages={messages}
          currentTranscript=""
          isRecording={state === "recording"}
          onPlayMessage={handlePlayMessage}
          speechSupported={isSupported}
        />

        <StateIndicator 
          state={state} 
          error={error} 
          speechSupported={isSupported}
          speechError={recordingError}
        />
      </div>
    </div>
  );
}
