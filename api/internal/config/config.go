package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Port                  string
	LogLevel              string
	SessionTimeoutMinutes int
	ContextDir            string
	MaxContextSummaries   int
	GitRecentDays         int
	CORSAllowedOrigins    string
	WorkspaceDir          string
	KokoroTTSPath         string
	KokoroTTSModelPath    string
	KokoroTTSVoicesPath   string
	KokoroTTSVoice        string
	KokoroTTSSpeed        float64
	WhisperPath           string
	WhisperModel          string
}

const (
	// DefaultPort is the default HTTP server port
	DefaultPort = "3000"
	// DefaultLogLevel is the default logging level
	DefaultLogLevel = "info"
	// DefaultSessionTimeoutMinutes is the default session timeout
	DefaultSessionTimeoutMinutes = 10
	// DefaultContextDir is the default context directory
	DefaultContextDir = ".janus"
	// DefaultMaxContextSummaries is the default number of summaries to load
	DefaultMaxContextSummaries = 3
	// DefaultGitRecentDays is the default number of days for recent files
	DefaultGitRecentDays = 3
	// DefaultCORSAllowedOrigins is the default CORS allowed origins for development
	// Use "*" to allow all origins (useful for development with mobile/Tailscale)
	DefaultCORSAllowedOrigins = "*"
	// DefaultWorkspaceDir is the default workspace directory for cursor-agent
	DefaultWorkspaceDir = "."
	// DefaultKokoroTTSPath is the default path to kokoro-tts executable (WSL)
	DefaultKokoroTTSPath = "/home/sean/kokoro-env/bin/kokoro-tts"
	// DefaultKokoroTTSModelPath is the default path to the model file
	DefaultKokoroTTSModelPath = "/home/sean/kokoro-v1.0.onnx"
	// DefaultKokoroTTSVoicesPath is the default path to the voices file
	DefaultKokoroTTSVoicesPath = "/home/sean/voices-v1.0.bin"
	// DefaultKokoroTTSVoice is the default voice for TTS
	DefaultKokoroTTSVoice = "af_sarah"
	// DefaultKokoroTTSSpeed is the default speech speed
	DefaultKokoroTTSSpeed = 1
	// DefaultWhisperPath is the default path to whisper executable
	DefaultWhisperPath = "/home/sean/whisper-local/.venv/bin/whisper"
	// DefaultWhisperModel is the default Whisper model to use
	DefaultWhisperModel = "base"
)

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (ignore error if it doesn't exist)
	_ = godotenv.Load()

	cfg := &Config{
		Port:                  getEnv("PORT", DefaultPort),
		LogLevel:              getEnv("LOG_LEVEL", DefaultLogLevel),
		SessionTimeoutMinutes: getEnvAsInt("SESSION_TIMEOUT_MINUTES", DefaultSessionTimeoutMinutes),
		ContextDir:            getEnv("CONTEXT_DIR", DefaultContextDir),
		MaxContextSummaries:   getEnvAsInt("MAX_CONTEXT_SUMMARIES", DefaultMaxContextSummaries),
		GitRecentDays:         getEnvAsInt("GIT_RECENT_DAYS", DefaultGitRecentDays),
		CORSAllowedOrigins:    getEnv("CORS_ALLOWED_ORIGINS", DefaultCORSAllowedOrigins),
		WorkspaceDir:          getEnv("WORKSPACE_DIR", DefaultWorkspaceDir),
		KokoroTTSPath:         getEnv("KOKORO_TTS_PATH", DefaultKokoroTTSPath),
		KokoroTTSModelPath:    getEnv("KOKORO_TTS_MODEL_PATH", DefaultKokoroTTSModelPath),
		KokoroTTSVoicesPath:   getEnv("KOKORO_TTS_VOICES_PATH", DefaultKokoroTTSVoicesPath),
		KokoroTTSVoice:        getEnv("KOKORO_TTS_VOICE", DefaultKokoroTTSVoice),
		KokoroTTSSpeed:        getEnvAsFloat("KOKORO_TTS_SPEED", DefaultKokoroTTSSpeed),
		WhisperPath:           getEnv("WHISPER_PATH", DefaultWhisperPath),
		WhisperModel:          getEnv("WHISPER_MODEL", DefaultWhisperModel),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("PORT cannot be empty")
	}

	if c.SessionTimeoutMinutes < 1 {
		return fmt.Errorf("SESSION_TIMEOUT_MINUTES must be at least 1")
	}

	return nil
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt reads an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvAsFloat reads an environment variable as float64 or returns a default value
func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}

	return value
}
