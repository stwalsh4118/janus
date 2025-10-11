import type {
  HealthResponse,
  StartSessionResponse,
  AskRequest,
  AskResponse,
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
}

// Create a singleton instance
const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3000";
export const apiClient = new JanusApiClient(apiUrl);

