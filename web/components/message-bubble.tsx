"use client";

import { User, Bot, Volume2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { Message } from "@/lib/types";

interface MessageBubbleProps {
  message: Message;
  onPlay: (content: string) => void;
}

export function MessageBubble({ message, onPlay }: MessageBubbleProps) {
  const isUser = message.role === "user";

  const getRelativeTime = (date: Date) => {
    const seconds = Math.floor((new Date().getTime() - date.getTime()) / 1000);

    if (seconds < 60) return "Just now";
    if (seconds < 3600) return `${Math.floor(seconds / 60)} mins ago`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)} hours ago`;
    return `${Math.floor(seconds / 86400)} days ago`;
  };

  return (
    <div className={`flex flex-col gap-2 w-full ${isUser ? "items-start md:items-end" : "items-start"}`}>
      <div className={`flex items-start gap-3 ${isUser ? "flex-row md:flex-row-reverse" : "flex-row"}`}>
        <div
          className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full ${
            isUser ? "bg-primary" : "bg-muted"
          }`}
        >
          {isUser ? (
            <User className="h-4 w-4 text-primary-foreground" />
          ) : (
            <Bot className="h-4 w-4 text-muted-foreground" />
          )}
        </div>
        <div className="flex-1 min-w-0">
          <p className={`whitespace-pre-wrap text-sm text-foreground leading-relaxed ${isUser ? "text-left md:text-right" : "text-left"}`}>
            {message.content}
          </p>
        </div>
      </div>
      <div className={`flex items-center gap-2 ${isUser ? "pl-11 md:pl-0 md:pr-11" : "pl-11"}`}>
        <span className="text-sm text-muted-foreground">{getRelativeTime(message.timestamp)}</span>
        {!isUser && (
          <Button
            variant="ghost"
            size="sm"
            className="h-7 gap-1 px-2 text-sm"
            onClick={() => onPlay(message.content)}
            data-interactive
          >
            <Volume2 className="h-3.5 w-3.5" />
            Play
          </Button>
        )}
      </div>
    </div>
  );
}

