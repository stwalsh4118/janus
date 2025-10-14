"use client";

import { Settings, ChevronDown, ChevronUp } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { TTSProvider } from "@/lib/types";

interface StatusBarProps {
  isConnected: boolean;
  activeSessions: number;
  version?: string;
  ttsProvider: TTSProvider;
  onTtsProviderChange: (provider: TTSProvider) => void;
  isDebugOpen: boolean;
  onDebugToggle: () => void;
}

export function StatusBar({
  isConnected,
  activeSessions,
  version = "0.1.0",
  ttsProvider,
  onTtsProviderChange,
  isDebugOpen,
  onDebugToggle,
}: StatusBarProps) {
  const getStatusColor = () => {
    if (isConnected) return "bg-success";
    return "bg-destructive";
  };

  const getStatusText = () => {
    if (isConnected) return "Connected";
    return "Disconnected";
  };

  const getTtsLabel = () => {
    switch (ttsProvider) {
      case "auto":
        return "Auto";
      case "kokoro":
        return "Kokoro";
      case "browser":
        return "Browser";
    }
  };

  const cycleTtsProvider = () => {
    const providers: TTSProvider[] = ["auto", "kokoro", "browser"];
    const currentIndex = providers.indexOf(ttsProvider);
    const nextProvider = providers[(currentIndex + 1) % providers.length];
    onTtsProviderChange(nextProvider);
  };

  return (
    <div
      className="relative z-20 flex items-center justify-between border-b border-border bg-background/95 px-4 py-3 backdrop-blur"
      data-interactive
    >
      <div className="flex items-center gap-2">
        <div className={`h-2 w-2 rounded-full ${getStatusColor()}`} />
        <span className="text-sm font-medium">{getStatusText()}</span>
      </div>

      <div className="flex items-center gap-2">
        <Button
          variant="ghost"
          size="sm"
          onClick={cycleTtsProvider}
          className="h-8 gap-2 text-xs"
          title="Change TTS Provider"
        >
          <Settings className="h-3 w-3" />
          TTS: {getTtsLabel()}
        </Button>

        <Button
          variant="ghost"
          size="sm"
          onClick={onDebugToggle}
          className="h-8 gap-1 text-xs"
          title="Toggle Debug Panel"
        >
          {isDebugOpen ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />}
          Debug
        </Button>

        <div className="text-xs text-muted-foreground">
          v{version} • {activeSessions} active
        </div>
      </div>
    </div>
  );
}

