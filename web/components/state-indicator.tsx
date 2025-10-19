"use client";

import { Mic, MicOff, Loader2, Volume2, AlertCircle } from "lucide-react";
import type { AppState } from "@/lib/types";

interface StateIndicatorProps {
  state: AppState;
  error?: string | null;
  speechSupported?: boolean;
  speechError?: string | null;
}

export function StateIndicator({ state, error, speechSupported = true, speechError }: StateIndicatorProps) {
  const getIndicatorContent = () => {
    switch (state) {
      case "recording":
        return {
          icon: <Mic className="h-[120px] w-[120px] text-destructive" />,
          label: "Recording...",
          helper: "Tap anywhere to stop recording",
        };
      case "processing":
        return {
          icon: <Loader2 className="h-[120px] w-[120px] animate-spin text-primary" />,
          label: "Processing...",
          helper: "Transcribing and getting response...",
        };
      case "speaking":
        return {
          icon: <Volume2 className="h-[120px] w-[120px] text-success" />,
          label: "Speaking...",
          helper: "Tap anywhere to stop playback",
        };
      case "error":
        return {
          icon: <AlertCircle className="h-[120px] w-[120px] text-destructive" />,
          label: "Error",
          helper: error || "Something went wrong",
        };
      case "idle":
      default:
        // Show different message if speech is not supported
        if (!speechSupported || speechError) {
          return {
            icon: <AlertCircle className="h-[120px] w-[120px] text-yellow-500" />,
            label: "Voice Input Unavailable",
            helper: "Voice requires HTTPS. Open Debug Panel (⚙️) to use text input instead",
          };
        }
        return {
          icon: <MicOff className="h-[120px] w-[120px] text-muted-foreground" />,
          label: "Tap Anywhere to Start",
          helper: "Touch anywhere on screen to start recording",
        };
    }
  };

  const { icon, label, helper } = getIndicatorContent();

  return (
    <div className="relative z-20 flex flex-col items-center justify-center py-12 pb-16 pointer-events-none">
      <div className="flex flex-col items-center gap-4">
        {icon}
        <div className="flex flex-col items-center gap-1">
          <p className="text-lg font-bold text-foreground">{label}</p>
          <p className={`text-center text-sm ${state === "error" ? "text-destructive" : "text-muted-foreground"}`}>
            {helper}
          </p>
        </div>
      </div>
    </div>
  );
}

