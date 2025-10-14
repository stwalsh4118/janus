import { useState, useEffect, useCallback, useRef } from 'react';
import { apiClient } from '@/lib/api-client';
import type { TTSProvider } from '@/lib/types';

interface UseSpeechSynthesisReturn {
  isSupported: boolean;
  isSpeaking: boolean;
  isGenerating: boolean;
  currentProvider: TTSProvider;
  preferredProvider: TTSProvider;
  setPreferredProvider: (provider: TTSProvider) => void;
  speak: (text: string) => Promise<void>;
  stop: () => void;
  pause: () => void;
  resume: () => void;
}

const STORAGE_KEY = 'janus-tts-provider';
const MAX_CACHE_SIZE = 10; // Cache audio for last 10 messages

export function useSpeechSynthesis(): UseSpeechSynthesisReturn {
  const [isSupported] = useState(true); // Always supported (browser fallback)
  const [isSpeaking, setIsSpeaking] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);
  const [currentProvider, setCurrentProvider] = useState<TTSProvider>('browser');
  const [preferredProvider, setPreferredProviderState] = useState<TTSProvider>('auto');
  const [kokoroAvailable, setKokoroAvailable] = useState<boolean>(false);
  const [voices, setVoices] = useState<SpeechSynthesisVoice[]>([]);
  
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const audioUrlRef = useRef<string | null>(null);
  const synthRef = useRef<SpeechSynthesis | null>(null);
  const audioCacheRef = useRef<Map<string, Blob>>(new Map());
  const shouldPlayRef = useRef<boolean>(true);

  // Initialize browser TTS
  useEffect(() => {
    if (typeof window !== 'undefined' && window.speechSynthesis) {
      synthRef.current = window.speechSynthesis;

      const loadVoices = () => {
        const availableVoices = window.speechSynthesis.getVoices();
        setVoices(availableVoices);
      };

      loadVoices();
      
      if (window.speechSynthesis.onvoiceschanged !== undefined) {
        window.speechSynthesis.onvoiceschanged = loadVoices;
      }
    }

    return () => {
      if (synthRef.current) {
        synthRef.current.cancel();
      }
    };
  }, []);

  // Check Kokoro availability on mount
  useEffect(() => {
    const checkKokoro = async () => {
      try {
        const health = await apiClient.checkTTSHealth();
        setKokoroAvailable(health.available && health.provider === 'kokoro');
      } catch (err) {
        console.warn('Failed to check Kokoro TTS availability:', err);
        setKokoroAvailable(false);
      }
    };

    checkKokoro();
  }, []);

  // Load preferred provider from localStorage
  useEffect(() => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(STORAGE_KEY) as TTSProvider | null;
      if (stored && ['auto', 'browser', 'kokoro'].includes(stored)) {
        setPreferredProviderState(stored);
      }
    }
  }, []);

  // Save preferred provider to localStorage
  const setPreferredProvider = useCallback((provider: TTSProvider) => {
    setPreferredProviderState(provider);
    if (typeof window !== 'undefined') {
      localStorage.setItem(STORAGE_KEY, provider);
    }
  }, []);

  const cleanupAudio = useCallback(() => {
    if (audioRef.current) {
      // Remove event listeners before cleanup to prevent error events
      audioRef.current.onplay = null;
      audioRef.current.onended = null;
      audioRef.current.onerror = null;
      
      audioRef.current.pause();
      audioRef.current.src = '';
      audioRef.current = null;
    }
    
    if (audioUrlRef.current) {
      URL.revokeObjectURL(audioUrlRef.current);
      audioUrlRef.current = null;
    }
  }, []);

  const cleanupBrowserTTS = useCallback(() => {
    if (synthRef.current) {
      synthRef.current.cancel();
    }
  }, []);

  const speakWithKokoro = useCallback(async (text: string) => {
    cleanupAudio();
    cleanupBrowserTTS();
    setCurrentProvider('kokoro');
    shouldPlayRef.current = true;

    try {
      // Check cache first
      let audioBlob = audioCacheRef.current.get(text);
      
      if (!audioBlob) {
        // Not in cache, generate new audio
        setIsGenerating(true);
        audioBlob = await apiClient.generateSpeech(text);
        
        // Check if stop was called during generation
        if (!shouldPlayRef.current) {
          setIsGenerating(false);
          return;
        }
        
        // Add to cache and manage cache size
        audioCacheRef.current.set(text, audioBlob);
        if (audioCacheRef.current.size > MAX_CACHE_SIZE) {
          // Remove oldest entry (first entry in the map)
          const firstKey = audioCacheRef.current.keys().next().value;
          if (firstKey) {
            audioCacheRef.current.delete(firstKey);
          }
        }
      }
      
      setIsGenerating(false);
      
      // Check again if stop was called
      if (!shouldPlayRef.current) {
        return;
      }
      
      const audioUrl = URL.createObjectURL(audioBlob);
      audioUrlRef.current = audioUrl;
      
      const audio = new Audio(audioUrl);
      audioRef.current = audio;

      audio.onplay = () => {
        setIsSpeaking(true);
      };

      audio.onended = () => {
        setIsSpeaking(false);
        cleanupAudio();
      };

      audio.onerror = (event) => {
        console.error('Kokoro audio playback error:', event);
        setIsSpeaking(false);
        cleanupAudio();
      };

      await audio.play();
    } catch (err) {
      console.error('Kokoro TTS failed:', err);
      setIsGenerating(false);
      cleanupAudio();
      throw err;
    }
  }, [cleanupAudio, cleanupBrowserTTS]);

  const speakWithBrowser = useCallback(async (text: string) => {
    cleanupAudio();
    cleanupBrowserTTS();
    setCurrentProvider('browser');

    if (!synthRef.current || !text) return;

    const utterance = new SpeechSynthesisUtterance(text);
    
    // Select best voice
    const preferredVoice = voices.find(
      voice => voice.lang === 'en-US' && voice.localService
    ) || voices.find(
      voice => voice.lang.startsWith('en')
    ) || voices[0];

    if (preferredVoice) {
      utterance.voice = preferredVoice;
    }

    utterance.rate = 1.0;
    utterance.pitch = 1.0;
    utterance.volume = 1.0;
    utterance.lang = 'en-US';

    utterance.onstart = () => {
      setIsSpeaking(true);
    };

    utterance.onend = () => {
      setIsSpeaking(false);
    };

    utterance.onerror = (event) => {
      console.error('Browser TTS error:', event);
      setIsSpeaking(false);
    };

    synthRef.current.speak(utterance);
  }, [voices, cleanupAudio, cleanupBrowserTTS]);

  const speak = useCallback(async (text: string) => {
    if (!text) return;

    // Determine which provider to use
    let providerToUse: 'browser' | 'kokoro' = 'browser';

    if (preferredProvider === 'auto') {
      // Auto mode: use Kokoro if available, otherwise browser
      providerToUse = kokoroAvailable ? 'kokoro' : 'browser';
    } else if (preferredProvider === 'kokoro') {
      // Force Kokoro
      providerToUse = 'kokoro';
    } else {
      // Force browser
      providerToUse = 'browser';
    }

    try {
      if (providerToUse === 'kokoro') {
        await speakWithKokoro(text);
      } else {
        await speakWithBrowser(text);
      }
    } catch (err) {
      // If Kokoro fails and we're in auto mode, fall back to browser
      if (providerToUse === 'kokoro' && preferredProvider === 'auto') {
        console.log('Kokoro TTS failed, falling back to browser TTS');
        await speakWithBrowser(text);
      } else {
        throw err;
      }
    }
  }, [preferredProvider, kokoroAvailable, speakWithKokoro, speakWithBrowser]);

  const stop = useCallback(() => {
    shouldPlayRef.current = false;
    cleanupAudio();
    cleanupBrowserTTS();
    setIsSpeaking(false);
    setIsGenerating(false);
  }, [cleanupAudio, cleanupBrowserTTS]);

  const pause = useCallback(() => {
    if (audioRef.current && !audioRef.current.paused) {
      audioRef.current.pause();
      setIsSpeaking(false);
    } else if (synthRef.current && synthRef.current.speaking) {
      synthRef.current.pause();
      setIsSpeaking(false);
    }
  }, []);

  const resume = useCallback(() => {
    if (audioRef.current && audioRef.current.paused) {
      audioRef.current.play().catch((err) => {
        console.error('Failed to resume audio:', err);
      });
      setIsSpeaking(true);
    } else if (synthRef.current && synthRef.current.paused) {
      synthRef.current.resume();
      setIsSpeaking(true);
    }
  }, []);

  return {
    isSupported,
    isSpeaking,
    isGenerating,
    currentProvider,
    preferredProvider,
    setPreferredProvider,
    speak,
    stop,
    pause,
    resume,
  };
}

