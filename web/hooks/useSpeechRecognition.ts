import { useState, useEffect, useCallback, useRef } from 'react';

interface UseSpeechRecognitionReturn {
  isSupported: boolean;
  isListening: boolean;
  transcript: string;
  interimTranscript: string;
  error: string | null;
  startListening: () => void;
  stopListening: () => void;
  resetTranscript: () => void;
  getCurrentTranscript: () => string;
}

export function useSpeechRecognition(): UseSpeechRecognitionReturn {
  const [isSupported, setIsSupported] = useState(false);
  const [isListening, setIsListening] = useState(false);
  const [transcript, setTranscript] = useState('');
  const [interimTranscript, setInterimTranscript] = useState('');
  const [error, setError] = useState<string | null>(null);
  
  const recognitionRef = useRef<SpeechRecognition | null>(null);
  
  // Refs to track latest values (avoid stale closures in async contexts)
  const transcriptRef = useRef<string>('');
  const interimTranscriptRef = useRef<string>('');
  
  // Keep refs in sync with state
  useEffect(() => {
    transcriptRef.current = transcript;
  }, [transcript]);
  
  useEffect(() => {
    interimTranscriptRef.current = interimTranscript;
  }, [interimTranscript]);

  useEffect(() => {
    // Check for browser support
    if (typeof window !== 'undefined') {
      const SpeechRecognitionAPI = 
        window.SpeechRecognition || window.webkitSpeechRecognition;
      
      if (SpeechRecognitionAPI) {
        setIsSupported(true);
        const recognition = new SpeechRecognitionAPI();
        
        // Configuration
        recognition.continuous = true;
        recognition.interimResults = true;
        recognition.lang = 'en-US';
        recognition.maxAlternatives = 1;

        // Event handlers
        recognition.onresult = (event: SpeechRecognitionEvent) => {
          let interimText = '';
          let finalText = '';

          for (let i = event.resultIndex; i < event.results.length; i++) {
            const result = event.results[i];
            const text = result[0].transcript;

            if (result.isFinal) {
              finalText += text + ' ';
            } else {
              interimText += text;
            }
          }

          if (finalText) {
            setTranscript(prev => prev + finalText);
          }
          setInterimTranscript(interimText);
        };

        recognition.onerror = (event: SpeechRecognitionErrorEvent) => {
          console.error('Speech recognition error:', event.error);
          setError(event.error);
          setIsListening(false);
        };

        recognition.onend = () => {
          setIsListening(false);
          setInterimTranscript('');
        };

        recognitionRef.current = recognition;
      }
    }

    return () => {
      if (recognitionRef.current) {
        recognitionRef.current.stop();
      }
    };
  }, []);

  const startListening = useCallback(() => {
    if (!recognitionRef.current || isListening) return;

    try {
      setError(null);
      recognitionRef.current.start();
      setIsListening(true);
    } catch (err) {
      console.error('Failed to start recognition:', err);
      setError('Failed to start speech recognition');
    }
  }, [isListening]);

  const stopListening = useCallback(() => {
    if (!recognitionRef.current || !isListening) return;

    try {
      recognitionRef.current.stop();
    } catch (err) {
      console.error('Failed to stop recognition:', err);
    }
  }, [isListening]);

  const resetTranscript = useCallback(() => {
    setTranscript('');
    setInterimTranscript('');
    setError(null);
  }, []);

  // Get the complete current transcript (final + interim)
  // This always returns the latest values, safe to use in async contexts
  const getCurrentTranscript = useCallback(() => {
    return (transcriptRef.current + interimTranscriptRef.current).trim();
  }, []);

  return {
    isSupported,
    isListening,
    transcript,
    interimTranscript,
    error,
    startListening,
    stopListening,
    resetTranscript,
    getCurrentTranscript,
  };
}


