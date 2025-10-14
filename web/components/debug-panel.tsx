"use client";

import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Volume2, Settings } from "lucide-react";
import type { TTSProvider } from "@/lib/types";

interface DebugPanelProps {
  isOpen: boolean;
  sessionId: string | null;
  debugText: string;
  onDebugTextChange: (text: string) => void;
  onDebugSubmit: () => void;
  isSending: boolean;
  ttsProvider: TTSProvider;
  currentProvider: TTSProvider;
  onTtsProviderChange: (provider: TTSProvider) => void;
}

export function DebugPanel({
  isOpen,
  sessionId,
  debugText,
  onDebugTextChange,
  onDebugSubmit,
  isSending,
  ttsProvider,
  currentProvider,
  onTtsProviderChange,
}: DebugPanelProps) {
  if (!isOpen) return null;

  const cycleTtsProvider = () => {
    const providers: TTSProvider[] = ["auto", "kokoro", "browser"];
    const currentIndex = providers.indexOf(ttsProvider);
    const nextProvider = providers[(currentIndex + 1) % providers.length];
    onTtsProviderChange(nextProvider);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" && debugText.trim() && !isSending && sessionId) {
      onDebugSubmit();
    }
  };

  return (
    <div className="relative z-20 border-b border-border bg-background/95 backdrop-blur" data-interactive>
      <div className="p-4 space-y-4">
        {/* TTS Provider Info */}
        <Card>
          <CardContent className="pt-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Volume2 className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-sm font-medium">
                    TTS: {ttsProvider === "auto" ? "Auto" : ttsProvider === "kokoro" ? "Kokoro" : "Browser"}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    Currently using: {currentProvider === "kokoro" ? "Kokoro AI" : currentProvider === "browser" ? "Browser" : "Auto"} |{" "}
                    {ttsProvider === "auto" && "Will use Kokoro if available, fallback to browser"}
                    {ttsProvider === "kokoro" && "Kokoro only mode"}
                    {ttsProvider === "browser" && "Browser only mode"}
                  </p>
                </div>
              </div>
              <Button variant="outline" size="sm" onClick={cycleTtsProvider}>
                <Settings className="h-4 w-4 mr-2" />
                Change
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Text Input */}
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Silent Mode: Text Input</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex gap-2">
              <Input
                type="text"
                placeholder="Type your question silently..."
                value={debugText}
                onChange={(e) => onDebugTextChange(e.target.value)}
                onKeyDown={handleKeyDown}
                disabled={isSending || !sessionId}
              />
              <Button onClick={onDebugSubmit} disabled={!debugText.trim() || isSending || !sessionId}>
                {isSending ? "Sending..." : "Ask"}
              </Button>
            </div>
            <p className="text-xs text-muted-foreground mt-2">
              Type your question and press Enter to ask without using voice
            </p>
          </CardContent>
        </Card>

        {/* Session Info */}
        {sessionId && (
          <Card>
            <CardContent className="pt-4">
              <p className="text-xs text-muted-foreground">
                Session ID: <code className="rounded bg-muted px-1 py-0.5">{sessionId}</code>
              </p>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}

