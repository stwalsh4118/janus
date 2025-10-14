package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sean/janus/internal/api/middleware"
	"github.com/sean/janus/internal/config"
	"github.com/sean/janus/internal/logger"
)

const (
	// TempFileCleanupAge is the minimum age before temp files can be deleted
	// Set to request timeout + 1 hour buffer to prevent deletion of files in use
	TempFileCleanupAge = middleware.DefaultRequestTimeout + 1*time.Hour
	// TempFileCleanupBuffer adds extra safety margin beyond request timeout
	TempFileCleanupBuffer = 1 * time.Hour
)

// TTSHandler handles text-to-speech generation requests
type TTSHandler struct {
	config *config.Config
}

// NewTTSHandler creates a new TTS handler
func NewTTSHandler(cfg *config.Config) *TTSHandler {
	return &TTSHandler{config: cfg}
}

// TTSRequest represents the request body for TTS generation
type TTSRequest struct {
	Text string `json:"text" binding:"required"`
}

// GenerateSpeech generates speech audio from text using kokoro-tts CLI
func (h *TTSHandler) GenerateSpeech(ctx context.Context, text string) (string, error) {
	log := logger.Get()

	// Create temp directory for TTS files if it doesn't exist
	tempDir := filepath.Join(os.TempDir(), "janus-tts")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Note: Cleanup is NOT done here to avoid race conditions with concurrent requests
	// Files are cleaned up by a background goroutine or by the handler after response is sent

	// Create unique temp files for input and output
	timestamp := time.Now().UnixNano()
	inputFile := filepath.Join(tempDir, fmt.Sprintf("input_%d.txt", timestamp))
	outputFile := filepath.Join(tempDir, fmt.Sprintf("output_%d.wav", timestamp))

	// Write text to temp file
	if err := os.WriteFile(inputFile, []byte(text), 0644); err != nil {
		return "", fmt.Errorf("failed to write input file: %w", err)
	}
	defer os.Remove(inputFile) // Clean up input file after generation

	// Execute kokoro-tts CLI (native WSL executable) with timeout from context
	cmd := exec.CommandContext(
		ctx,
		h.config.KokoroTTSPath,
		inputFile,
		outputFile,
		"--model", h.config.KokoroTTSModelPath,
		"--voices", h.config.KokoroTTSVoicesPath,
		"--speed", fmt.Sprintf("%.1f", h.config.KokoroTTSSpeed),
		"--lang", "en-us",
		"--voice", h.config.KokoroTTSVoice,
	)

	// Set environment variable for GPU acceleration
	cmd.Env = append(os.Environ(), "ONNX_PROVIDER=CUDAExecutionProvider")

	log.Debug().
		Str("kokoro_path", h.config.KokoroTTSPath).
		Str("model_path", h.config.KokoroTTSModelPath).
		Str("voices_path", h.config.KokoroTTSVoicesPath).
		Str("voice", h.config.KokoroTTSVoice).
		Float64("speed", h.config.KokoroTTSSpeed).
		Str("input_file", inputFile).
		Str("output_file", outputFile).
		Str("onnx_provider", "CUDAExecutionProvider").
		Msg("Executing kokoro-tts command")

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if error was due to context cancellation (timeout)
		if ctx.Err() == context.DeadlineExceeded {
			log.Error().
				Err(err).
				Str("output", string(output)).
				Msg("kokoro-tts command timed out")
			return "", fmt.Errorf("kokoro-tts timed out: %w", err)
		}
		log.Error().
			Err(err).
			Str("output", string(output)).
			Msg("kokoro-tts command failed")
		return "", fmt.Errorf("kokoro-tts failed: %w\nOutput: %s", err, output)
	}

	log.Debug().
		Str("output", string(output)).
		Msg("kokoro-tts command succeeded")

	return outputFile, nil
}

// cleanupOldTempFiles removes temp files older than the specified age threshold
// The threshold should be large enough to avoid deleting files from concurrent requests
func (h *TTSHandler) cleanupOldTempFiles(tempDir string, ageThreshold time.Duration) {
	log := logger.Get()

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to read temp directory for cleanup")
		return
	}

	now := time.Now()
	cleanupCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Delete files older than the age threshold
		if now.Sub(info.ModTime()) > ageThreshold {
			filePath := filepath.Join(tempDir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				log.Warn().
					Err(err).
					Str("file", filePath).
					Msg("Failed to remove old temp file")
			} else {
				cleanupCount++
			}
		}
	}

	if cleanupCount > 0 {
		log.Debug().
			Int("count", cleanupCount).
			Dur("age_threshold", ageThreshold).
			Msg("Cleaned up old TTS temp files")
	}
}

// Generate handles the HTTP request for TTS generation
func (h *TTSHandler) Generate(c *gin.Context) {
	log := logger.Get()

	var req TTSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid TTS request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Text cannot be empty"})
		return
	}

	log.Info().
		Int("text_length", len(req.Text)).
		Msg("Generating TTS audio")

	// Perform background cleanup of old temp files (safe from race conditions)
	tempDir := filepath.Join(os.TempDir(), "janus-tts")
	go h.cleanupOldTempFiles(tempDir, TempFileCleanupAge)

	// Generate speech audio with context (includes timeout from middleware)
	audioPath, err := h.GenerateSpeech(c.Request.Context(), req.Text)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate speech")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate speech"})
		return
	}

	// Ensure the audio file is cleaned up after sending
	defer func() {
		if err := os.Remove(audioPath); err != nil {
			log.Warn().
				Err(err).
				Str("file", audioPath).
				Msg("Failed to remove audio file after sending")
		}
	}()

	// Stream the WAV file as response
	c.Header("Content-Type", "audio/wav")
	c.File(audioPath)

	log.Info().Msg("TTS audio sent successfully")
}
