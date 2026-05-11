package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// measures how long it took req to complete
func LatencyCheckerMiddlewareFunction() gin.HandlerFunc {
	return func(c *gin.Context) {

		invocationTime := time.Now()
		
		c.Next()

		latency := (time.Since(invocationTime))
		slog.Info("time took to complete request","time :",latency)

	}
}