package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Logger middleware
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// RateLimiter middleware (basic implementation)
func RateLimiter() gin.HandlerFunc {
	// This is a simple rate limiter for demonstration
	// In production, you would use a more sophisticated solution like Redis
	clientRequests := make(map[string][]time.Time)
	
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()
		
		// Clean old requests (older than 1 minute)
		if requests, exists := clientRequests[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < time.Minute {
					validRequests = append(validRequests, reqTime)
				}
			}
			clientRequests[clientIP] = validRequests
		}
		
		// Check if client has exceeded rate limit (60 requests per minute)
		if len(clientRequests[clientIP]) >= 60 {
			c.JSON(429, gin.H{"error": "Rate limit exceeded. Please try again later."})
			c.Abort()
			return
		}
		
		// Add current request
		clientRequests[clientIP] = append(clientRequests[clientIP], now)
		
		c.Next()
	}
}