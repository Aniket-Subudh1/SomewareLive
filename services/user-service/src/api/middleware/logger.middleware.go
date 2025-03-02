package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Logger is a middleware for logging HTTP requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Get request ID
		requestID := c.Writer.Header().Get("X-Request-ID")
		if requestID == "" {
			requestID = c.GetString("request_id")
		}

		// Get authenticated user ID if available
		userID := GetUserId(c)

		// Calculate latency
		latency := time.Since(start)

		// Get client IP
		clientIP := c.ClientIP()

		// Get status
		status := c.Writer.Status()

		// Get user agent
		userAgent := c.Request.UserAgent()

		// Get error if any
		errs := c.Errors.ByType(gin.ErrorTypePrivate)
		errMsg := ""
		if len(errs) > 0 {
			errMsg = errs[0].Error()
		}

		// Construct query path
		if raw != "" {
			path = path + "?" + raw
		}

		// Log request
		logEvent := log.Info()

		// Add fields
		logEvent = logEvent.
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", status).
			Dur("latency", latency).
			Str("ip", clientIP).
			Str("user-agent", userAgent)

		// Add request ID if available
		if requestID != "" {
			logEvent = logEvent.Str("request_id", requestID)
		}

		// Add user ID if available
		if userID != "" {
			logEvent = logEvent.Str("user_id", userID)
		}

		// Add error if any
		if errMsg != "" {
			logEvent = logEvent.Str("error", errMsg)
		}

		// Determine log level based on status code
		if status >= 500 {
			log.Error().
				Str("method", c.Request.Method).
				Str("path", path).
				Int("status", status).
				Dur("latency", latency).
				Str("ip", clientIP).
				Str("user-agent", userAgent).
				Str("error", errMsg).
				Msg("Server error")
		} else if status >= 400 {
			log.Warn().
				Str("method", c.Request.Method).
				Str("path", path).
				Int("status", status).
				Dur("latency", latency).
				Str("ip", clientIP).
				Str("user-agent", userAgent).
				Str("error", errMsg).
				Msg("Client error")
		} else {
			logEvent.Msg("Request")
		}
	}
}

// RequestID is a middleware for adding a request ID to the context
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request ID from header
		requestID := c.Request.Header.Get("X-Request-ID")

		// If not present, generate a new one
		if requestID == "" {
			requestID = GenerateRequestID()
		}

		// Set request ID in context
		c.Set("request_id", requestID)

		// Set request ID in response header
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}

// GenerateRequestID generates a unique request ID
func GenerateRequestID() string {
	// Simple implementation using timestamp
	return time.Now().Format("20060102150405") + "-" + RandomString(8)
}

// RandomString generates a random string of the specified length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[RandomInt(len(charset))]
	}
	return string(b)
}

// RandomInt generates a random integer between 0 and n-1
func RandomInt(n int) int {
	return int(time.Now().UnixNano() % int64(n))
}
