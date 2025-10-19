package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sean/janus/internal/config"
	"github.com/sean/janus/internal/logger"
)

// TranscribeHandler handles audio transcription requests
type TranscribeHandler struct {
	config *config.Config
}

// NewTranscribeHandler creates a new transcribe handler
func NewTranscribeHandler(cfg *config.Config) *TranscribeHandler {
	return &TranscribeHandler{
		config: cfg,
	}
}

// TranscribeResponse represents the transcription response
type TranscribeResponse struct {
	Text string `json:"text"`
}

// Transcribe processes audio transcription requests
func (h *TranscribeHandler) Transcribe(c *gin.Context) {
	log := logger.Get()

	// Get the uploaded audio file
	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get audio file from request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No audio file provided"})
		return
	}
	defer file.Close()

	log.Info().
		Str("filename", header.Filename).
		Int64("size", header.Size).
		Msg("Received audio file for transcription")

	// Create temp directory for audio processing
	tempDir := filepath.Join(os.TempDir(), "janus-transcribe")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create temp directory")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Create temp file for audio input
	timestamp := time.Now().UnixNano()
	audioExt := filepath.Ext(header.Filename)
	if audioExt == "" {
		audioExt = ".webm" // Default for browser recordings
	}
	audioPath := filepath.Join(tempDir, fmt.Sprintf("audio_%d%s", timestamp, audioExt))

	// Save uploaded file
	audioFile, err := os.Create(audioPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create audio file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if _, err := io.Copy(audioFile, file); err != nil {
		audioFile.Close()
		os.Remove(audioPath)
		log.Error().Err(err).Msg("Failed to save audio file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	audioFile.Close()

	// Clean up audio file after processing
	defer os.Remove(audioPath)

	// Run Whisper transcription with timeout
	text, err := h.runWhisper(c, audioPath)
	if err != nil {
		log.Error().Err(err).Msg("Whisper transcription failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transcription failed"})
		return
	}

	// Log success at Info level (without PII), transcription text at Debug level only
	log.Info().Msg("Transcription successful")
	log.Debug().
		Str("text", text).
		Msg("Transcription text")

	c.JSON(http.StatusOK, TranscribeResponse{
		Text: text,
	})
}

// runWhisper executes the Whisper command and returns the transcribed text
func (h *TranscribeHandler) runWhisper(c *gin.Context, audioPath string) (string, error) {
	log := logger.Get()

	// Build whisper command
	// whisper audio.webm --model base --output_format txt --output_dir /tmp
	outputDir := filepath.Dir(audioPath)

	// Create context with timeout (2 minutes should be enough for most audio clips)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		h.config.WhisperPath,
		audioPath,
		"--model", h.config.WhisperModel,
		"--output_format", "txt",
		"--output_dir", outputDir,
	)

	log.Debug().
		Str("whisper_path", h.config.WhisperPath).
		Str("audio_path", audioPath).
		Str("model", h.config.WhisperModel).
		Str("output_dir", outputDir).
		Msg("Executing whisper command")

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if timeout occurred
		if ctx.Err() == context.DeadlineExceeded {
			log.Error().
				Str("output", string(output)).
				Msg("Whisper command timed out after 2 minutes")
			return "", fmt.Errorf("whisper command timed out: %w", ctx.Err())
		}

		log.Error().
			Err(err).
			Str("output", string(output)).
			Msg("Whisper command failed")
		return "", fmt.Errorf("whisper command failed: %w", err)
	}

	log.Debug().
		Str("output", string(output)).
		Msg("Whisper command succeeded")

	// Read the generated .txt file
	baseName := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
	txtPath := filepath.Join(outputDir, baseName+".txt")
	defer os.Remove(txtPath) // Clean up the .txt file

	textBytes, err := os.ReadFile(txtPath)
	if err != nil {
		log.Error().
			Err(err).
			Str("txt_path", txtPath).
			Msg("Failed to read transcription file")
		return "", fmt.Errorf("failed to read transcription: %w", err)
	}

	text := strings.TrimSpace(string(textBytes))
	return text, nil
}
