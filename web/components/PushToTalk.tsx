"use client";

import { Button } from "@/components/ui/button";
import { Mic, MicOff } from "lucide-react";
import type { PushToTalkProps } from "@/lib/types";

export function PushToTalk({ disabled, isRecording, isSending, onToggle }: PushToTalkProps) {
  const handleClick = () => {
    if (disabled) return;
    onToggle();
  };

  // Determine button text and helper text based on state
  const getButtonText = () => {
    if (isSending) return "Sending...";
    if (disabled) return "Disabled";
    if (isRecording) return "Tap to Stop";
    return "Tap to Talk";
  };

  const getHelperText = () => {
    if (isSending) return "Waiting for response...";
    if (disabled) return "Connect to backend to start";
    return "Tap once to start recording, tap again to send";
  };

  return (
    <div className="flex flex-col items-center gap-4 py-8">
      <Button
        size="lg"
        disabled={disabled}
        onClick={handleClick}
        className={`
          h-48 w-48 rounded-full transition-all duration-200
          ${isRecording ? "scale-95 shadow-inner" : "scale-100 shadow-lg"}
          ${disabled ? "opacity-50 cursor-not-allowed" : "cursor-pointer"}
        `}
        variant={isRecording ? "secondary" : "default"}
      >
        <div className="flex flex-col items-center gap-2">
          {isRecording ? (
            <Mic className="h-16 w-16 animate-pulse" />
          ) : (
            <MicOff className="h-16 w-16" />
          )}
          <span className="text-sm font-medium">
            {getButtonText()}
          </span>
        </div>
      </Button>
      
      <p className="text-xs text-muted-foreground text-center max-w-xs">
        {getHelperText()}
      </p>
    </div>
  );
}

