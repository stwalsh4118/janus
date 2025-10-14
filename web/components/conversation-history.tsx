"use client";

import { useEffect, useRef } from "react";
import { User } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { MessageBubble } from "@/components/message-bubble";
import type { Message } from "@/lib/types";

interface ConversationHistoryProps {
  messages: Message[];
  currentTranscript: string;
  isRecording: boolean;
  onPlayMessage: (content: string) => void;
}

export function ConversationHistory({
  messages,
  currentTranscript,
  isRecording,
  onPlayMessage,
}: ConversationHistoryProps) {
  const scrollViewportRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (scrollViewportRef.current) {
      const viewport = scrollViewportRef.current.querySelector("[data-radix-scroll-area-viewport]") as HTMLDivElement;
      if (viewport) {
        viewport.scrollTop = viewport.scrollHeight;
      }
    }
  }, [messages, currentTranscript]);

  return (
    <div className="relative z-10 min-h-0 flex-1" ref={scrollViewportRef}>
      <ScrollArea className="h-full">
        <div className="mx-auto max-w-6xl pointer-events-auto" data-interactive>
          <div className="flex flex-col gap-6 p-4">
          {messages.length === 0 && !currentTranscript && (
            <div className="flex h-full items-center justify-center py-20">
              <p className="text-center text-sm text-muted-foreground">Tap anywhere to ask a question</p>
            </div>
          )}

            {messages.map((message) => (
              <MessageBubble key={message.id} message={message} onPlay={onPlayMessage} />
            ))}

          {currentTranscript && isRecording && (
            <div className="flex flex-col gap-2 w-full">
              <div className="flex items-start gap-3">
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary">
                  <User className="h-4 w-4 text-primary-foreground" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="whitespace-pre-wrap text-sm text-foreground leading-relaxed animate-pulse">{currentTranscript}</p>
                </div>
              </div>
              <div className="flex items-center gap-2 pl-11">
                <span className="text-sm text-muted-foreground">Recording...</span>
              </div>
            </div>
          )}
          </div>
        </div>
      </ScrollArea>
    </div>
  );
}

