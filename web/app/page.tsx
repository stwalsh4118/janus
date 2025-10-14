"use client";

import { useEffect, useState } from "react";
import { StatusIndicator } from "@/components/StatusIndicator";
import { PushToTalk } from "@/components/PushToTalk";
import { SpeechUnsupported } from "@/components/SpeechUnsupported";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Input } from "@/components/ui/input";
import { Volume2, VolumeX, Settings } from "lucide-react";
import { apiClient } from "@/lib/api-client";
import { useSpeechRecognition } from "@/hooks/useSpeechRecognition";
import { useSpeechSynthesis } from "@/hooks/useSpeechSynthesis";
import type { HealthResponse } from "@/lib/types";

export default function Home() {
  const [status, setStatus] = useState<"connected" | "disconnected" | "connecting">("connecting");
  const [healthData, setHealthData] = useState<HealthResponse | null>(null);
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [response, setResponse] = useState<string>("");
  const [error, setError] = useState<string>("");
  const [isSending, setIsSending] = useState<boolean>(false);
  const [debugText, setDebugText] = useState<string>("");

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

  // Check backend health on mount
  useEffect(() => {
    const checkHealth = async () => {
      try {
        setStatus("connecting");
        const data = await apiClient.healthCheck();
        setHealthData(data);
        setStatus("connected");
        setError("");
      } catch (err) {
        console.error("Health check failed:", err);
        setStatus("disconnected");
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
    if (status === "connected" && !sessionId) {
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
  }, [status, sessionId]);

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
      // This is more reliable than async fetch during unload
      const url = `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:3000"}/api/session/end?session_id=${encodeURIComponent(sessionId)}`;
      
      if (navigator.sendBeacon) {
        navigator.sendBeacon(url);
      } else {
        // Fallback for browsers that don't support sendBeacon
        // Note: This may not complete if the page unloads quickly
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

  const handleToggleTalk = async () => {
    if (isListening) {
      // Prevent duplicate sends
      if (isSending) return;
      
      setIsSending(true);
      
      await new Promise(resolve => setTimeout(resolve, 500));
      
      stopListening();
      
      await new Promise(resolve => setTimeout(resolve, 150));
      
      const questionToSend = getCurrentTranscript();
      
      console.log("=== Speech Debug ===");
      console.log("Combined (sending):", questionToSend);
      console.log("===================");
      
      if (!sessionId || !questionToSend) {
        // If no transcript, just reset
        resetTranscript();
        setIsSending(false);
        return;
      }

      try {
        setError(""); // Clear any previous errors
        const answer = await apiClient.ask(sessionId, questionToSend);
        setResponse(answer);
        
        // Auto-play response with TTS
        if (ttsSupported && answer) {
          speak(answer);
        }
        
        // Reset transcript after successful send
        resetTranscript();
      } catch (err) {
        console.error("Failed to ask question:", err);
        setError(err instanceof Error ? err.message : "Failed to get response");
      } finally {
        setIsSending(false);
      }
    } else {
      // Start recording
      resetTranscript();
      startListening();
    }
  };

  // Combined transcript for display (final + interim)
  const displayTranscript = transcript + interimTranscript;

  return (
    <main className="min-h-screen bg-background p-4 md:p-8">
      <div className="mx-auto max-w-4xl space-y-6">
        {/* Header */}
        <div className="text-center">
          <h1 className="text-4xl font-bold tracking-tight">Janus</h1>
          <p className="mt-2 text-muted-foreground">
            Your voice portal to the codebase
          </p>
        </div>

        {/* Status */}
        <StatusIndicator
          status={status}
          version={healthData?.version}
          activeSessions={healthData?.active_sessions}
        />

        {/* TTS Provider Status */}
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Volume2 className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-sm font-medium">
                    TTS: {preferredProvider === 'auto' ? 'Auto' : preferredProvider === 'kokoro' ? 'Kokoro' : 'Browser'}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {preferredProvider === 'auto' && 'Will use Kokoro if available, fallback to browser'}
                    {preferredProvider === 'kokoro' && 'Kokoro only mode'}
                    {preferredProvider === 'browser' && 'Browser only mode'}
                  </p>
                </div>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  const providers: Array<'auto' | 'kokoro' | 'browser'> = ['auto', 'kokoro', 'browser'];
                  const currentIndex = providers.indexOf(preferredProvider);
                  const nextProvider = providers[(currentIndex + 1) % providers.length];
                  setPreferredProvider(nextProvider);
                }}
              >
                <Settings className="h-4 w-4 mr-2" />
                Change
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Speech Unsupported Warning */}
        {!isSupported && <SpeechUnsupported />}

        {/* Debug Text Input */}
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Debug: Text Input (Silent Mode)</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex gap-2">
              <Input
                type="text"
                placeholder="Type your question silently..."
                value={debugText}
                onChange={(e) => setDebugText(e.target.value)}
                onKeyDown={async (e) => {
                  if (e.key === 'Enter' && debugText.trim() && !isSending && sessionId) {
                    const question = debugText.trim();
                    setDebugText("");
                    setIsSending(true);
                    
                    try {
                      setError("");
                      const answer = await apiClient.ask(sessionId, question);
                      setResponse(answer);
                      
                      // Auto-play response
                      if (ttsSupported && answer) {
                        await speak(answer);
                      }
                    } catch (err) {
                      console.error("Failed to ask question:", err);
                      setError(err instanceof Error ? err.message : "Failed to get response");
                    } finally {
                      setIsSending(false);
                    }
                  }
                }}
                disabled={isSending || !sessionId}
              />
              <Button
                onClick={async () => {
                  if (debugText.trim() && !isSending && sessionId) {
                    const question = debugText.trim();
                    setDebugText("");
                    setIsSending(true);
                    
                    try {
                      setError("");
                      const answer = await apiClient.ask(sessionId, question);
                      setResponse(answer);
                      
                      // Auto-play response
                      if (ttsSupported && answer) {
                        await speak(answer);
                      }
                    } catch (err) {
                      console.error("Failed to ask question:", err);
                      setError(err instanceof Error ? err.message : "Failed to get response");
                    } finally {
                      setIsSending(false);
                    }
                  }
                }}
                disabled={!debugText.trim() || isSending || !sessionId}
              >
                {isSending ? "Sending..." : "Ask"}
              </Button>
            </div>
            <p className="text-xs text-muted-foreground mt-2">
              Type your question and press Enter to ask without using voice
            </p>
          </CardContent>
        </Card>

        {/* Error Display */}
        {(error || speechError) && (
          <Card className="border-destructive">
            <CardContent className="pt-6">
              <p className="text-sm text-destructive">{error || speechError}</p>
            </CardContent>
          </Card>
        )}

        {/* Push to Talk */}
        <div className="flex justify-center">
          <PushToTalk
            disabled={status !== "connected" || !sessionId || !isSupported || isSending}
            isRecording={isListening}
            isSending={isSending}
            onToggle={handleToggleTalk}
          />
        </div>

        {/* Transcript */}
        {displayTranscript && (
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">
                {isListening ? "Listening..." : isSending ? "Sending..." : "Transcript"}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground">
                {displayTranscript}
              </p>
            </CardContent>
          </Card>
        )}

        {/* Response */}
        {response && (
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <CardTitle className="text-lg">
                    {isGeneratingAudio ? "Generating audio..." : "Response"}
                  </CardTitle>
                  <span className="text-xs text-muted-foreground">
                    ({currentProvider === 'kokoro' ? 'Kokoro AI' : 'Browser'} TTS)
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  {(isSpeaking || isGeneratingAudio) && (
                    <Button 
                      variant="outline" 
                      size="sm"
                      onClick={stopSpeaking}
                      disabled={isGeneratingAudio}
                    >
                      <VolumeX className="h-4 w-4 mr-2" />
                      {isGeneratingAudio ? "Generating..." : "Stop"}
                    </Button>
                  )}
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => {
                      const providers: Array<'auto' | 'kokoro' | 'browser'> = ['auto', 'kokoro', 'browser'];
                      const currentIndex = providers.indexOf(preferredProvider);
                      const nextProvider = providers[(currentIndex + 1) % providers.length];
                      setPreferredProvider(nextProvider);
                    }}
                    title="Toggle TTS Provider"
                  >
                    <Settings className="h-4 w-4" />
                  </Button>
                </div>
              </div>
              <div className="text-xs text-muted-foreground mt-1">
                Mode: {preferredProvider === 'auto' ? 'Auto (Kokoro preferred)' : preferredProvider === 'kokoro' ? 'Kokoro only' : 'Browser only'}
              </div>
            </CardHeader>
            <CardContent>
              <div className="flex items-start gap-3">
                {(isSpeaking || isGeneratingAudio) && (
                  <Volume2 className={`h-5 w-5 text-primary ${isSpeaking ? 'animate-pulse' : 'animate-spin'}`} />
                )}
                <ScrollArea className="h-[200px] w-full rounded-md">
                  <p className="text-sm whitespace-pre-wrap">{response}</p>
                </ScrollArea>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Session Info */}
        {sessionId && (
          <Card>
            <CardContent className="pt-6">
              <p className="text-xs text-muted-foreground">
                Session ID: <code className="rounded bg-muted px-1 py-0.5">{sessionId}</code>
              </p>
            </CardContent>
          </Card>
        )}
      </div>
    </main>
  );
}
