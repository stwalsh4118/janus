package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sean/janus/internal/logger"
)

// TTSHealthResponse represents the TTS health check response
type TTSHealthResponse struct {
	Available bool   `json:"available"`
	Provider  string `json:"provider"`
	Voice     string `json:"voice,omitempty"`
	Message   string `json:"message,omitempty"`
}

// HealthCheck checks if Kokoro TTS is properly configured and available
func (h *TTSHandler) HealthCheck(c *gin.Context) {
	log := logger.Get()

	// Check if kokoro-tts executable exists
	if _, err := os.Stat(h.config.KokoroTTSPath); err != nil {
		if os.IsNotExist(err) {
			log.Debug().
				Str("path", h.config.KokoroTTSPath).
				Msg("Kokoro TTS executable not found")

			c.JSON(http.StatusOK, TTSHealthResponse{
				Available: false,
				Provider:  "browser",
				Message:   "Kokoro TTS not configured, using browser TTS",
			})
			return
		}
		// Handle other errors (permission, I/O, etc.)
		log.Error().
			Err(err).
			Str("path", h.config.KokoroTTSPath).
			Msg("Failed to check Kokoro TTS executable")

		c.JSON(http.StatusOK, TTSHealthResponse{
			Available: false,
			Provider:  "browser",
			Message:   fmt.Sprintf("Kokoro TTS inaccessible: %v", err),
		})
		return
	}

	// Check if model files exist
	if _, err := os.Stat(h.config.KokoroTTSModelPath); err != nil {
		if os.IsNotExist(err) {
			log.Debug().
				Str("path", h.config.KokoroTTSModelPath).
				Msg("Kokoro model file not found")

			c.JSON(http.StatusOK, TTSHealthResponse{
				Available: false,
				Provider:  "browser",
				Message:   "Kokoro model files not found, using browser TTS",
			})
			return
		}
		// Handle other errors (permission, I/O, etc.)
		log.Error().
			Err(err).
			Str("path", h.config.KokoroTTSModelPath).
			Msg("Failed to check Kokoro model file")

		c.JSON(http.StatusOK, TTSHealthResponse{
			Available: false,
			Provider:  "browser",
			Message:   fmt.Sprintf("Kokoro TTS inaccessible: %v", err),
		})
		return
	}

	if _, err := os.Stat(h.config.KokoroTTSVoicesPath); err != nil {
		if os.IsNotExist(err) {
			log.Debug().
				Str("path", h.config.KokoroTTSVoicesPath).
				Msg("Kokoro voices file not found")

			c.JSON(http.StatusOK, TTSHealthResponse{
				Available: false,
				Provider:  "browser",
				Message:   "Kokoro voices file not found, using browser TTS",
			})
			return
		}
		// Handle other errors (permission, I/O, etc.)
		log.Error().
			Err(err).
			Str("path", h.config.KokoroTTSVoicesPath).
			Msg("Failed to check Kokoro voices file")

		c.JSON(http.StatusOK, TTSHealthResponse{
			Available: false,
			Provider:  "browser",
			Message:   fmt.Sprintf("Kokoro TTS inaccessible: %v", err),
		})
		return
	}

	// All checks passed - Kokoro TTS is available
	log.Debug().Msg("Kokoro TTS is available and configured")

	c.JSON(http.StatusOK, TTSHealthResponse{
		Available: true,
		Provider:  "kokoro",
		Voice:     h.config.KokoroTTSVoice,
		Message:   "Kokoro TTS available",
	})
}
