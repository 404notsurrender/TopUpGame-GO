package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger creates a gin middleware for logging HTTP requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Stop timer
		duration := time.Since(start)

		// Get request details
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		status := c.Writer.Status()
		size := c.Writer.Size()

		if query != "" {
			path = path + "?" + query
		}

		// Get client IP
		clientIP := c.ClientIP()

		// Get error if any
		if len(c.Errors) > 0 {
			// Log error
			fmt.Printf("[ERROR] %v | %3d | %13v | %15s | %-7s %s | %s\n",
				start.Format("2006/01/02 - 15:04:05"),
				status,
				duration,
				clientIP,
				method,
				path,
				c.Errors.String(),
			)
		} else {
			// Log request
			fmt.Printf("[INFO] %v | %3d | %13v | %15s | %-7s %s | %d bytes\n",
				start.Format("2006/01/02 - 15:04:05"),
				status,
				duration,
				clientIP,
				method,
				path,
				size,
			)
		}
	}
}

// ErrorLogger creates a gin middleware for logging errors
func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Log only if there are errors
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				fmt.Printf("[ERROR] %v | Error: %s | Path: %s\n",
					time.Now().Format("2006/01/02 - 15:04:05"),
					e.Error(),
					c.Request.URL.Path,
				)
			}
		}
	}
}

// RecoveryLogger creates a gin middleware for recovering from panics and logging them
func RecoveryLogger() gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultErrorWriter, func(c *gin.Context, err interface{}) {
		fmt.Printf("[PANIC] %v | Recovered from panic: %v | Path: %s\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			err,
			c.Request.URL.Path,
		)
		c.AbortWithStatus(500)
	})
}
