"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { StatusBar } from "@/components/status-bar";
import { ConversationHistory } from "@/components/conversation-history";
import { StateIndicator } from "@/components/state-indicator";
import { DebugPanel } from "@/components/debug-panel";
import { apiClient } from "@/lib/api-client";
import { useSpeechRecognition } from "@/hooks/useSpeechRecognition";
import { useSpeechSynthesis } from "@/hooks/useSpeechSynthesis";
import type { AppState, Message, HealthResponse } from "@/lib/types";

export default function Home() {
  const [state, setState] = useState<AppState>("idle");
  const [messages, setMessages] = useState<Message[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [healthData, setHealthData] = useState<HealthResponse | null>(null);
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [isDebugOpen, setIsDebugOpen] = useState(false);
  const [debugText, setDebugText] = useState<string>("");

  const containerRef = useRef<HTMLDivElement>(null);

  // Speech recognition hook
  const {
    isSupported,
    isListening,
    transcript,
    interimTranscript,
    error: speechError,
    startListening,
    stopListening,
    resetTranscript,
    getCurrentTranscript,
  } = useSpeechRecognition();

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
  } = useSpeechSynthesis();

  // Combined transcript for display (final + interim)
  const displayTranscript = transcript + interimTranscript;

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
      const url = `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:3000"}/api/session/end?session_id=${encodeURIComponent(sessionId)}`;

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

  // Update error state based on speech error
  useEffect(() => {
    if (speechError && state !== "error") {
      setError(speechError);
      setState("error");
    }
  }, [speechError, state]);

  // Update state based on speech and TTS status
  useEffect(() => {
    if (isSpeaking || isGeneratingAudio) {
      setState("speaking");
    } else if (isListening) {
      setState("recording");
    } else if (state === "speaking" || state === "recording") {
      // Only reset to idle if we were speaking or recording
      setState("idle");
    }
  }, [isSpeaking, isGeneratingAudio, isListening, state]);

  const processQuestion = useCallback(
    async (question: string) => {
      if (!sessionId || !question) return;

      setState("processing");

      try {
        setError(null);

        // Add user message
        const userMessage: Message = {
          id: crypto.randomUUID(),
          role: "user",
          content: question,
          timestamp: new Date(),
        };
        setMessages((prev) => [...prev, userMessage]);

        // Get answer
        const answer = await apiClient.ask(sessionId, question);

        // Add assistant message
        const assistantMessage: Message = {
          id: crypto.randomUUID(),
          role: "assistant",
          content: answer,
          timestamp: new Date(),
        };
        setMessages((prev) => [...prev, assistantMessage]);

        // Auto-play response with TTS
        if (ttsSupported && answer) {
          await speak(answer);
        } else {
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
      if (state === "speaking") {
        stopSpeaking();
      }
      await speak(content);
    },
    [state, speak, stopSpeaking]
  );

  const startRecording = useCallback(() => {
    if (!isSupported) {
      setError("Speech recognition is not supported in your browser");
      setState("error");
      return;
    }

    resetTranscript();
    startListening();
    setState("recording");
    setError(null);
  }, [isSupported, resetTranscript, startListening]);

  const stopRecording = useCallback(async () => {
    stopListening();

    // Wait a bit for final transcript
    await new Promise((resolve) => setTimeout(resolve, 500));

    const questionToSend = getCurrentTranscript();

    if (!questionToSend) {
      resetTranscript();
      setState("idle");
      return;
    }

    // Process the question
    await processQuestion(questionToSend);

    // Reset transcript after successful send
    resetTranscript();
  }, [stopListening, getCurrentTranscript, resetTranscript, processQuestion]);

  const handleDebugSubmit = useCallback(async () => {
    if (!debugText.trim() || !sessionId) return;

    const question = debugText.trim();
    setDebugText("");

    await processQuestion(question);
  }, [debugText, sessionId, processQuestion]);

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
        startRecording();
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
          currentTranscript={displayTranscript}
          isRecording={state === "recording"}
          onPlayMessage={handlePlayMessage}
        />

        <StateIndicator state={state} error={error} />
      </div>
    </div>
  );
}
