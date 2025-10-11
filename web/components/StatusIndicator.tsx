"use client";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import type { StatusIndicatorProps } from "@/lib/types";

export function StatusIndicator({
  status,
  version,
  activeSessions,
}: StatusIndicatorProps) {
  const statusConfig = {
    connected: {
      label: "Connected",
      variant: "default" as const,
      color: "bg-green-500",
    },
    disconnected: {
      label: "Disconnected",
      variant: "destructive" as const,
      color: "bg-red-500",
    },
    connecting: {
      label: "Connecting...",
      variant: "secondary" as const,
      color: "bg-yellow-500",
    },
  };

  const config = statusConfig[status];

  return (
    <Card className="w-full">
      <CardContent className="pt-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className={`h-3 w-3 rounded-full ${config.color} animate-pulse`} />
            <div>
              <p className="text-sm font-medium">Backend Status</p>
              <Badge variant={config.variant} className="mt-1">
                {config.label}
              </Badge>
            </div>
          </div>
          
          {status === "connected" && (
            <div className="text-right">
              {version && (
                <p className="text-xs text-muted-foreground">
                  Version {version}
                </p>
              )}
              {activeSessions !== undefined && (
                <p className="text-xs text-muted-foreground mt-1">
                  {activeSessions} active session{activeSessions !== 1 ? "s" : ""}
                </p>
              )}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

