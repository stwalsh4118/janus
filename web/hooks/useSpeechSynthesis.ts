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
  unlockAudio: () => void;
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
  const audioUnlockedRef = useRef<boolean>(false);
  const preparedAudioRef = useRef<HTMLAudioElement | null>(null); // Pre-created audio for user gesture
  const audioContextRef = useRef<AudioContext | null>(null);
  const audioSourceRef = useRef<AudioBufferSourceNode | null>(null);
  const debugIndicatorRef = useRef<HTMLDivElement | null>(null);

  // Initialize browser TTS and Web Audio API
  useEffect(() => {
    if (typeof window !== 'undefined') {
      // Initialize Web Audio API for better iOS compatibility
      // Note: On iOS, AudioContext might start in 'suspended' state until user interaction
      if ('AudioContext' in window || 'webkitAudioContext' in window) {
        const AudioContextClass = (window.AudioContext || (window as any).webkitAudioContext);
        audioContextRef.current = new AudioContextClass();
        console.log('[Audio] Web Audio API initialized, state:', audioContextRef.current.state, 'sampleRate:', audioContextRef.current.sampleRate);
      } else {
        console.warn('[Audio] Web Audio API not supported in this browser');
      }
      
      if (window.speechSynthesis) {
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
    }

    return () => {
      if (synthRef.current) {
        synthRef.current.cancel();
      }
      if (audioContextRef.current) {
        console.log('[Audio] Closing AudioContext on unmount');
        audioContextRef.current.close().catch(console.error);
      }
      // Clean up debug indicator if component unmounts
      if (debugIndicatorRef.current) {
        debugIndicatorRef.current.remove();
        debugIndicatorRef.current = null;
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
    // Stop Web Audio API source if playing
    if (audioSourceRef.current) {
      try {
        audioSourceRef.current.stop();
      } catch (e) {
        // Already stopped
      }
      audioSourceRef.current = null;
    }
    
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

  // Play audio using Web Audio API (better iOS compatibility)
  const playWithWebAudio = useCallback(async (audioBlob: Blob): Promise<boolean> => {
    if (!audioContextRef.current) {
      console.warn('[WebAudio] AudioContext not available');
      return false;
    }

    try {
      console.log('[WebAudio] AudioContext state BEFORE:', audioContextRef.current.state);
      console.log('[WebAudio] AudioContext sampleRate:', audioContextRef.current.sampleRate);
      
      // Resume audio context if suspended (iOS requirement)
      if (audioContextRef.current.state === 'suspended') {
        console.log('[WebAudio] Resuming suspended AudioContext...');
        await audioContextRef.current.resume();
        console.log('[WebAudio] AudioContext state AFTER resume:', audioContextRef.current.state);
      }
      
      console.log('[WebAudio] Converting blob to ArrayBuffer...');
      const arrayBuffer = await audioBlob.arrayBuffer();
      console.log('[WebAudio] ArrayBuffer size:', arrayBuffer.byteLength);
      
      console.log('[WebAudio] Decoding audio data...');
      const audioBuffer = await audioContextRef.current.decodeAudioData(arrayBuffer);
      console.log('[WebAudio] Audio decoded successfully!');
      console.log('[WebAudio] - duration:', audioBuffer.duration);
      console.log('[WebAudio] - channels:', audioBuffer.numberOfChannels);
      console.log('[WebAudio] - sampleRate:', audioBuffer.sampleRate);
      console.log('[WebAudio] - length:', audioBuffer.length);
      
      // Create gain node for volume control
      console.log('[WebAudio] Creating gain node...');
      const gainNode = audioContextRef.current.createGain();
      gainNode.gain.value = 1.0;
      console.log('[WebAudio] Gain value set to:', gainNode.gain.value);
      
      console.log('[WebAudio] Creating audio source...');
      const source = audioContextRef.current.createBufferSource();
      source.buffer = audioBuffer;
      
      // Connect: source -> gain -> destination
      source.connect(gainNode);
      gainNode.connect(audioContextRef.current.destination);
      console.log('[WebAudio] Audio chain connected: source -> gain -> destination');
      
      // Set up event handlers
      source.onended = () => {
        console.log('[WebAudio] Playback ended naturally');
        setIsSpeaking(false);
        audioSourceRef.current = null;
        
        // Remove debug indicator
        if (debugIndicatorRef.current) {
          debugIndicatorRef.current.remove();
          debugIndicatorRef.current = null;
        }
      };
      
      audioSourceRef.current = source;
      setIsSpeaking(true);
      
      console.log('[WebAudio] Starting playback NOW...');
      console.log('[WebAudio] AudioContext state:', audioContextRef.current.state);
      source.start(0);
      console.log('[WebAudio] source.start() called successfully!');
      console.log('[WebAudio] AudioContext.currentTime:', audioContextRef.current.currentTime);
      
      // Remove any existing debug indicator
      if (debugIndicatorRef.current) {
        debugIndicatorRef.current.remove();
      }
      
      // Add visual debug indicator
      const debugDiv = document.createElement('div');
      debugDiv.id = 'audio-debug';
      debugDiv.style.cssText = 'position:fixed;top:10px;right:10px;z-index:99999;background:lime;color:black;padding:10px;font-weight:bold;border-radius:8px;font-size:14px;';
      debugDiv.textContent = `🔊 PLAYING (${audioBuffer.duration.toFixed(1)}s)`;
      document.body.appendChild(debugDiv);
      debugIndicatorRef.current = debugDiv;
      
      return true;
    } catch (err) {
      console.error('[WebAudio] Failed to play audio:', err);
      console.error('[WebAudio] Error details:', JSON.stringify(err, Object.getOwnPropertyNames(err)));
      return false;
    }
  }, []);

  // Unlock audio for Safari autoplay
  // This creates a silent audio context during user interaction
  // which allows subsequent programmatic audio playback
  const unlockAudio = useCallback(() => {
    if (audioUnlockedRef.current) return;

    try {
      // Resume AudioContext if suspended (required for iOS)
      if (audioContextRef.current && audioContextRef.current.state === 'suspended') {
        console.log('[Audio] Resuming AudioContext...');
        audioContextRef.current.resume().then(() => {
          console.log('[Audio] AudioContext resumed, state:', audioContextRef.current?.state);
        });
      }
      
      // Create a silent audio buffer (0.1 seconds of silence)
      const silentAudio = new Audio('data:audio/mp3;base64,SUQzBAAAAAAAI1RTU0UAAAAPAAADTGF2ZjU4Ljc2LjEwMAAAAAAAAAAAAAAA//tQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAWGluZwAAAA8AAAACAAADhAC7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7u7//////////////////////////////////////////////////////////////////8AAAAATGF2YzU4LjEzAAAAAAAAAAAAAAAAJAAAAAAAAAAAA4R+AAAAAAAAAAAAAAAAAAAAAP/7kGQAD/AAAGkAAAAIAAANIAAAAQAAAaQAAAAgAAA0gAAABExBTUUzLjEwMFVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVf/7kGQAD/AAAGkAAAAIAAANIAAAAQAAAaQAAAAgAAA0gAAABFVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVQ==');
      silentAudio.play().then(() => {
        audioUnlockedRef.current = true;
        console.log('[Audio] Unlocked for autoplay');
      }).catch((err) => {
        console.warn('[Audio] Failed to unlock:', err);
      });
    } catch (err) {
      console.warn('[Audio] Failed to create silent audio for unlock:', err);
    }

    // Also pre-create an Audio element to maintain user gesture context
    if (!preparedAudioRef.current) {
      console.log('[Audio] Pre-creating Audio element for next playback');
      preparedAudioRef.current = new Audio();
    }
  }, []);

  const speakWithKokoro = useCallback(async (text: string) => {
    console.log('[Kokoro] speakWithKokoro called with text:', text.substring(0, 50) + '...');
    cleanupAudio();
    cleanupBrowserTTS();
    setCurrentProvider('kokoro');
    shouldPlayRef.current = true;

    try {
      // Try to use pre-created audio element if available (maintains user gesture context)
      let audio = preparedAudioRef.current;
      if (audio) {
        console.log('[Kokoro] Using pre-created Audio element from user gesture');
        preparedAudioRef.current = null; // Clear it so it's not reused
      } else {
        console.log('[Kokoro] Creating new Audio element...');
        audio = new Audio();
      }
      
      audioRef.current = audio;
      console.log('[Kokoro] Audio element ready:', audio);

      // Force unmute and set volume (Safari sometimes mutes by default)
      audio.muted = false;
      audio.volume = 1.0;
      console.log('[Kokoro] Audio explicitly unmuted and volume set to 1.0');

      // Set up event handlers before loading
      audio.onplay = () => {
        console.log('[Kokoro] onplay event fired');
        setIsSpeaking(true);
      };

      audio.onloadedmetadata = () => {
        console.log('[Kokoro] onloadedmetadata - duration:', audio.duration, 'volume:', audio.volume, 'muted:', audio.muted);
      };

      audio.oncanplay = () => {
        console.log('[Kokoro] oncanplay - ready to play');
      };

      audio.ontimeupdate = () => {
        console.log('[Kokoro] ontimeupdate - currentTime:', audio.currentTime, '/', audio.duration);
      };

      audio.onended = () => {
        console.log('[Kokoro] onended event fired');
        setIsSpeaking(false);
        cleanupAudio();
      };

      audio.onerror = (event) => {
        console.error('[Kokoro] audio playback error:', event);
        console.error('[Kokoro] audio.error:', audio.error);
        setIsSpeaking(false);
        cleanupAudio();
      };

      audio.onpause = () => {
        console.log('[Kokoro] onpause event fired');
      };

      audio.onstalled = () => {
        console.warn('[Kokoro] onstalled - download stalled');
      };

      // Check cache first
      let audioBlob = audioCacheRef.current.get(text);
      console.log('[Kokoro] Cache check:', audioBlob ? 'HIT' : 'MISS');
      
      if (!audioBlob) {
        // Not in cache, generate new audio
        console.log('[Kokoro] Generating speech from API...');
        setIsGenerating(true);
        audioBlob = await apiClient.generateSpeech(text);
        console.log('[Kokoro] Audio blob received:', audioBlob.size, 'bytes');
        
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
      
      // iOS Safari hack: Use HTML audio element with blob URL
      // Web Audio API works on desktop but iOS silences it
      console.log('[Kokoro] Setting up HTML audio playback...');
      console.log('[Kokoro] Creating audio URL from blob...');
      const audioUrl = URL.createObjectURL(audioBlob);
      audioUrlRef.current = audioUrl;
      audio.src = audioUrl;
      console.log('[Kokoro] Audio URL set:', audioUrl);
      console.log('[Kokoro] Audio blob type:', audioBlob.type, 'size:', audioBlob.size);

      // Load and play the audio
      console.log('[Kokoro] Loading audio...');
      audio.load();
      
      // Set up event handlers for debugging
      audio.onplay = () => {
        console.log('[Kokoro] Audio PLAY event fired');
        setIsSpeaking(true);
        
        // Add visual debug indicator
        if (debugIndicatorRef.current) {
          debugIndicatorRef.current.remove();
        }
        const debugDiv = document.createElement('div');
        debugDiv.style.cssText = 'position:fixed;top:10px;right:10px;z-index:99999;background:lime;color:black;padding:10px;font-weight:bold;border-radius:8px;font-size:14px;';
        debugDiv.textContent = `🔊 PLAYING`;
        document.body.appendChild(debugDiv);
        debugIndicatorRef.current = debugDiv;
      };
      
      audio.onended = () => {
        console.log('[Kokoro] Audio ENDED event fired');
        setIsSpeaking(false);
        if (debugIndicatorRef.current) {
          debugIndicatorRef.current.remove();
          debugIndicatorRef.current = null;
        }
      };
      
      audio.onpause = () => {
        console.log('[Kokoro] Audio PAUSE event fired');
      };
      
      audio.onerror = (e) => {
        console.error('[Kokoro] Audio ERROR event fired:', e);
      };
      
      try {
        console.log('[Kokoro] Calling audio.play()...');
        const playPromise = audio.play();
        console.log('[Kokoro] audio.play() returned promise');
        
        await playPromise;
        console.log('[Kokoro] Play promise resolved!');
        console.log('[Kokoro] Audio state - paused:', audio.paused, 'currentTime:', audio.currentTime, 'duration:', audio.duration, 'volume:', audio.volume, 'muted:', audio.muted, 'readyState:', audio.readyState);
      } catch (playErr) {
        console.error('[Kokoro] Play failed:', playErr);
        console.error('[Kokoro] Error name:', (playErr as Error).name);
        console.error('[Kokoro] Error message:', (playErr as Error).message);
        setIsSpeaking(false);
      }
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

    // Unlock audio on first speak attempt (for Safari autoplay)
    unlockAudio();

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
  }, [preferredProvider, kokoroAvailable, speakWithKokoro, speakWithBrowser, unlockAudio]);

  const stop = useCallback(() => {
    console.log('[TTS] Stop called');
    shouldPlayRef.current = false;
    
    // Remove debug indicator
    if (debugIndicatorRef.current) {
      debugIndicatorRef.current.remove();
      debugIndicatorRef.current = null;
    }
    
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
    unlockAudio,
  };
}

