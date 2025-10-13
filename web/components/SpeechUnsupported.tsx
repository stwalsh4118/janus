"use client";

import { Card, CardContent } from "@/components/ui/card";
import { AlertCircle } from "lucide-react";

export function SpeechUnsupported() {
  return (
    <Card className="border-yellow-500">
      <CardContent className="pt-6">
        <div className="flex items-start gap-3">
          <AlertCircle className="h-5 w-5 text-yellow-500" />
          <div>
            <h3 className="font-semibold">Speech Recognition Not Supported</h3>
            <p className="text-sm text-muted-foreground mt-1">
              Your browser doesn&apos;t support speech recognition. 
              Please use Chrome, Edge, or Safari on iOS.
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}


