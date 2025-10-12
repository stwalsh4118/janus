// Web Speech API type definitions
// Supports both standard and webkit-prefixed APIs for Safari compatibility

declare global {
  interface Window {
    // Safari uses webkit prefix
    webkitSpeechRecognition: typeof SpeechRecognition;
    SpeechRecognition: typeof SpeechRecognition;
  }

  // Enhanced SpeechRecognitionEvent interface
  interface SpeechRecognitionEvent extends Event {
    readonly results: SpeechRecognitionResultList;
    readonly resultIndex: number;
    readonly interpretation: any;
    readonly emma: Document | null;
  }

  // SpeechRecognitionErrorEvent interface
  interface SpeechRecognitionErrorEvent extends Event {
    readonly error: 
      | 'no-speech'
      | 'aborted'
      | 'audio-capture'
      | 'network'
      | 'not-allowed'
      | 'service-not-allowed'
      | 'bad-grammar'
      | 'language-not-supported';
    readonly message: string;
  }
}

export {};

