// API Response Types

export interface HealthResponse {
  status: string;
  version: string;
  uptime_seconds: number;
  active_sessions: number;
}

export interface StartSessionResponse {
  session_id: string;
  message: string;
}

export interface AskRequest {
  question: string;
}

export interface AskResponse {
  answer: string;
}

export interface GenericResponse {
  success: boolean;
  message: string;
}

export interface ErrorResponse {
  error: string;
}

export interface TTSHealthResponse {
  available: boolean;
  provider: "browser" | "kokoro";
  voice?: string;
  message?: string;
}

export type TTSProvider = "browser" | "kokoro" | "auto";

// Component Props Types

export interface StatusIndicatorProps {
  status: "connected" | "disconnected" | "connecting";
  version?: string;
  activeSessions?: number;
}

export interface PushToTalkProps {
  disabled?: boolean;
  isRecording: boolean;
  isSending?: boolean;
  onToggle: () => void;
}

// State Types

export interface AppState {
  sessionId: string | null;
  isConnected: boolean;
  backendStatus: HealthResponse | null;
  transcript: string;
  response: string;
  isRecording: boolean;
  isProcessing: boolean;
}

