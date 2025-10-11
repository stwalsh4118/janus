"use client";

import { useEffect, useState } from "react";
import { StatusIndicator } from "@/components/StatusIndicator";
import { PushToTalk } from "@/components/PushToTalk";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { apiClient } from "@/lib/api-client";
import type { HealthResponse } from "@/lib/types";

export default function Home() {
  const [status, setStatus] = useState<"connected" | "disconnected" | "connecting">("connecting");
  const [healthData, setHealthData] = useState<HealthResponse | null>(null);
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [transcript, setTranscript] = useState<string>("");
  const [response, setResponse] = useState<string>("");
  const [error, setError] = useState<string>("");

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

  const handlePressToTalk = () => {
    setTranscript("Recording... (stub - voice recognition will be added in PBI-4)");
  };

  const handleReleaseToTalk = async () => {
    if (!sessionId) return;

    const stubQuestion = "How does the authentication work in this project?";
    setTranscript(`You asked: "${stubQuestion}"`);
    
    try {
      const answer = await apiClient.ask(sessionId, stubQuestion);
      setResponse(answer);
    } catch (err) {
      console.error("Failed to ask question:", err);
      setError(err instanceof Error ? err.message : "Failed to get response");
    }
  };

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

        {/* Error Display */}
        {error && (
          <Card className="border-destructive">
            <CardContent className="pt-6">
              <p className="text-sm text-destructive">{error}</p>
            </CardContent>
          </Card>
        )}

        {/* Push to Talk */}
        <div className="flex justify-center">
          <PushToTalk
            disabled={status !== "connected" || !sessionId}
            onPress={handlePressToTalk}
            onRelease={handleReleaseToTalk}
          />
        </div>

        {/* Transcript */}
        {transcript && (
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Transcript</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground">{transcript}</p>
            </CardContent>
          </Card>
        )}

        {/* Response */}
        {response && (
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Response</CardTitle>
            </CardHeader>
            <CardContent>
              <ScrollArea className="h-[200px] w-full rounded-md">
                <p className="text-sm whitespace-pre-wrap">{response}</p>
              </ScrollArea>
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
