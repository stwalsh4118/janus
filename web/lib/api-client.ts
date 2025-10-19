import type {
  HealthResponse,
  StartSessionResponse,
  AskRequest,
  AskResponse,
  TTSHealthResponse,
} from "./types";

export class JanusApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  /**
   * Check backend health status
   */
  async healthCheck(): Promise<HealthResponse> {
    const response = await fetch(`${this.baseUrl}/api/health`);
    
    if (!response.ok) {
      throw new Error(`Health check failed: ${response.statusText}`);
    }
    
    return response.json();
  }

  /**
   * Start a new chat session
   */
  async startSession(): Promise<string> {
    const response = await fetch(`${this.baseUrl}/api/session/start`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });
    
    if (!response.ok) {
      throw new Error(`Failed to start session: ${response.statusText}`);
    }
    
    const data: StartSessionResponse = await response.json();
    return data.session_id;
  }

  /**
   * Ask a question in the current session
   */
  async ask(sessionId: string, question: string): Promise<string> {
    const response = await fetch(
      `${this.baseUrl}/api/ask?session_id=${encodeURIComponent(sessionId)}`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ question } as AskRequest),
      }
    );
    
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || `Failed to ask question: ${response.statusText}`);
    }
    
    const data: AskResponse = await response.json();
    return data.answer;
  }

  /**
   * Send heartbeat to keep session alive
   */
  async heartbeat(sessionId: string): Promise<void> {
    const response = await fetch(
      `${this.baseUrl}/api/heartbeat?session_id=${encodeURIComponent(sessionId)}`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      }
    );
    
    if (!response.ok) {
      throw new Error(`Heartbeat failed: ${response.statusText}`);
    }
  }

  /**
   * End the current session
   */
  async endSession(sessionId: string): Promise<void> {
    const response = await fetch(
      `${this.baseUrl}/api/session/end?session_id=${encodeURIComponent(sessionId)}`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      }
    );
    
    if (!response.ok) {
      throw new Error(`Failed to end session: ${response.statusText}`);
    }
  }

  /**
   * Check if server-side TTS (Kokoro) is available
   */
  async checkTTSHealth(): Promise<TTSHealthResponse> {
    const response = await fetch(`${this.baseUrl}/api/tts/health`);
    
    if (!response.ok) {
      throw new Error(`TTS health check failed: ${response.statusText}`);
    }
    
    return response.json();
  }

  /**
   * Generate speech audio from text using Kokoro TTS
   */
  async generateSpeech(text: string): Promise<Blob> {
    const response = await fetch(`${this.baseUrl}/api/tts`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ text }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to generate speech: ${errorText || response.statusText}`);
    }

    return response.blob();
  }

  /**
   * Transcribe audio to text using Whisper
   */
  async transcribe(audioBlob: Blob): Promise<string> {
    const formData = new FormData();
    formData.append("audio", audioBlob, "recording.webm");

    const response = await fetch(`${this.baseUrl}/api/transcribe`, {
      method: "POST",
      body: formData,
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Transcription failed: ${errorText || response.statusText}`);
    }

    const data = await response.json();
    return data.text;
  }
}

// Create a singleton instance
// Use empty string for relative URLs (same origin) when using Tailscale path routing
// Or set NEXT_PUBLIC_API_URL for separate backend
const apiUrl = process.env.NEXT_PUBLIC_API_URL || "";
console.log('[API Client] Using API URL:', apiUrl || '(relative/same origin)');
console.log('[API Client] Environment variable:', process.env.NEXT_PUBLIC_API_URL);
export const apiClient = new JanusApiClient(apiUrl);

