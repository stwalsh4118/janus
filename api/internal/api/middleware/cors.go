package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSConfig creates a CORS middleware configuration
func CORSConfig(allowedOrigins string) gin.HandlerFunc {
	config := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	// Support wildcard "*" for development (allow all origins)
	if allowedOrigins == "*" {
		config.AllowAllOrigins = true
		config.AllowCredentials = false // Can't use credentials with AllowAllOrigins
	} else {
		// Support comma-separated list of origins
		origins := strings.Split(allowedOrigins, ",")
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
		config.AllowOrigins = origins
	}

	return cors.New(config)
}
