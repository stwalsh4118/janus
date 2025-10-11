"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Mic, MicOff } from "lucide-react";
import type { PushToTalkProps } from "@/lib/types";

export function PushToTalk({ disabled, onPress, onRelease }: PushToTalkProps) {
  const [isPressed, setIsPressed] = useState(false);

  const handleMouseDown = () => {
    if (disabled) return;
    setIsPressed(true);
    onPress?.();
  };

  const handleMouseUp = () => {
    if (disabled) return;
    setIsPressed(false);
    onRelease?.();
  };

  const handleTouchStart = (e: React.TouchEvent) => {
    e.preventDefault();
    if (disabled) return;
    setIsPressed(true);
    onPress?.();
  };

  const handleTouchEnd = (e: React.TouchEvent) => {
    e.preventDefault();
    if (disabled) return;
    setIsPressed(false);
    onRelease?.();
  };

  return (
    <div className="flex flex-col items-center gap-4 py-8">
      <Button
        size="lg"
        disabled={disabled}
        onMouseDown={handleMouseDown}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
        onTouchStart={handleTouchStart}
        onTouchEnd={handleTouchEnd}
        className={`
          h-48 w-48 rounded-full transition-all duration-200
          ${isPressed ? "scale-95 shadow-inner" : "scale-100 shadow-lg"}
          ${disabled ? "opacity-50 cursor-not-allowed" : "cursor-pointer"}
        `}
        variant={isPressed ? "secondary" : "default"}
      >
        <div className="flex flex-col items-center gap-2">
          {isPressed ? (
            <Mic className="h-16 w-16 animate-pulse" />
          ) : (
            <MicOff className="h-16 w-16" />
          )}
          <span className="text-sm font-medium">
            {disabled ? "Disabled" : isPressed ? "Recording..." : "Hold to Talk"}
          </span>
        </div>
      </Button>
      
      <p className="text-xs text-muted-foreground text-center max-w-xs">
        {disabled
          ? "Connect to backend to start"
          : "Press and hold the button to record your question"}
      </p>
    </div>
  );
}

