package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info().
			Str("request_id", c.GetString("request_id")).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency_ms", time.Since(start)).
			Str("client_ip", c.ClientIP()).
			Msg("request")
	}
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Error().Interface("panic", recovered).Msg("recovered from panic")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	})
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
		c.Next()
	}
}

func CORS(env string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if env == "development" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {

			c.Header("Access-Control-Allow-Origin", "https://app.nexus-logistics.com")
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

type bucket struct {
	tokens   float64
	lastSeen time.Time
	mu       sync.Mutex
}

var (
	buckets sync.Map
)

func startBucketCleanup(window time.Duration) {
	ticker := time.NewTicker(5 * time.Minute)
	go bucketCleanupLoop(window, ticker.C)
}

func bucketCleanupLoop(window time.Duration, tick <-chan time.Time) {
	for range tick {
		buckets.Range(func(k, v any) bool {
			b := v.(*bucket)
			b.mu.Lock()
			stale := time.Since(b.lastSeen) > 5*window
			b.mu.Unlock()
			if stale {
				buckets.Delete(k)
			}
			return true
		})
	}
}

func RateLimit(limit float64, window time.Duration) gin.HandlerFunc {
	refillRate := limit / window.Seconds()
	startBucketCleanup(window)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		val, _ := buckets.LoadOrStore(ip, &bucket{tokens: limit, lastSeen: now})
		b := val.(*bucket)

		b.mu.Lock()
		elapsed := now.Sub(b.lastSeen).Seconds()
		b.tokens = minFloat(limit, b.tokens+elapsed*refillRate)
		b.lastSeen = now

		if b.tokens < 1 {
			b.mu.Unlock()
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		b.tokens--
		b.mu.Unlock()

		c.Next()
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
