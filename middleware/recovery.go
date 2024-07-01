package middleware

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "log"
)

func RecoveryMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("Panic recovered: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
                c.Abort()
            }
        }()
        c.Next()
    }
}
