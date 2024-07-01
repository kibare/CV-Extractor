package middleware

import (
    "time"
    "github.com/gin-gonic/gin"
    "log"
)

func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        startTime := time.Now()
        c.Next()
        duration := time.Since(startTime)

        log.Printf("Method: %s, Path: %s, Status: %d, Duration: %s",
            c.Request.Method,
            c.Request.URL.Path,
            c.Writer.Status(),
            duration,
        )
    }
}
