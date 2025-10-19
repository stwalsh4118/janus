import { useState, useEffect, useRef, useCallback } from 'react';
import { apiClient } from '@/lib/api-client';

interface UseAudioRecorderReturn {
  isSupported: boolean;
  isRecording: boolean;
  isTranscribing: boolean;
  error: string | null;
  startRecording: () => void;
  stopRecording: () => Promise<string>;
  resetRecording: () => void;
}

export function useAudioRecorder(): UseAudioRecorderReturn {
  const [isSupported, setIsSupported] = useState(false);
  const [isRecording, setIsRecording] = useState(false);
  const [isTranscribing, setIsTranscribing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const audioChunksRef = useRef<Blob[]>([]);
  const streamRef = useRef<MediaStream | null>(null);

  // Check browser support on mount
  useEffect(() => {
    if (typeof window !== 'undefined' && navigator.mediaDevices && navigator.mediaDevices.getUserMedia) {
      setIsSupported(true);
    }
  }, []);

  const startRecording = useCallback(async () => {
    if (isRecording) return;

    try {
      setError(null);
      audioChunksRef.current = [];

      // Request microphone access
      const stream = await navigator.mediaDevices.getUserMedia({ 
        audio: {
          channelCount: 1,
          sampleRate: 16000, // Whisper prefers 16kHz
        } 
      });
      streamRef.current = stream;

      // Create MediaRecorder
      const mimeType = MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
        ? 'audio/webm;codecs=opus'
        : 'audio/webm';

      const mediaRecorder = new MediaRecorder(stream, { mimeType });
      mediaRecorderRef.current = mediaRecorder;

      // Collect audio chunks
      mediaRecorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          audioChunksRef.current.push(event.data);
        }
      };

      mediaRecorder.onerror = (event) => {
        console.error('MediaRecorder error:', event);
        setError('Recording error occurred');
        setIsRecording(false);
      };

      // Start recording
      mediaRecorder.start(100); // Collect data every 100ms
      setIsRecording(true);
      console.log('Audio recording started');
    } catch (err) {
      console.error('Failed to start recording:', err);
      const errorMessage = err instanceof Error 
        ? err.message 
        : 'Failed to start audio recording. Please allow microphone access.';
      setError(errorMessage);
    }
  }, [isRecording]);

  const stopRecording = useCallback(async (): Promise<string> => {
    if (!mediaRecorderRef.current || !isRecording) {
      return '';
    }

    return new Promise((resolve, reject) => {
      const mediaRecorder = mediaRecorderRef.current!;

      mediaRecorder.onstop = async () => {
        console.log('Audio recording stopped, processing...');
        
        // Stop all tracks
        if (streamRef.current) {
          streamRef.current.getTracks().forEach(track => track.stop());
          streamRef.current = null;
        }

        setIsRecording(false);

        // Create audio blob
        const audioBlob = new Blob(audioChunksRef.current, { type: 'audio/webm' });
        console.log('Audio blob created:', audioBlob.size, 'bytes');

        // Transcribe the audio
        try {
          setIsTranscribing(true);
          const transcription = await apiClient.transcribe(audioBlob);
          console.log('Transcription complete:', transcription);
          setIsTranscribing(false);
          setError(null);
          resolve(transcription);
        } catch (err) {
          console.error('Transcription failed:', err);
          setIsTranscribing(false);
          const errorMessage = err instanceof Error 
            ? err.message 
            : 'Failed to transcribe audio';
          setError(errorMessage);
          reject(new Error(errorMessage));
        }
      };

      mediaRecorder.stop();
    });
  }, [isRecording]);

  const resetRecording = useCallback(() => {
    if (mediaRecorderRef.current && isRecording) {
      mediaRecorderRef.current.stop();
    }
    if (streamRef.current) {
      streamRef.current.getTracks().forEach(track => track.stop());
      streamRef.current = null;
    }
    audioChunksRef.current = [];
    setIsRecording(false);
    setIsTranscribing(false);
    setError(null);
  }, [isRecording]);

  return {
    isSupported,
    isRecording,
    isTranscribing,
    error,
    startRecording,
    stopRecording,
    resetRecording,
  };
}

