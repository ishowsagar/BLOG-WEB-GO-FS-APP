package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// bug -> header overiding problems
func SlogLoggerMiddlewareFunction() gin.HandlerFunc {

	return func(c *gin.Context) {
		currentTime := time.Now()
		reqPath := c.Request.URL.Path
		queryParam := c.Request.URL.RawQuery
		
		var logger *slog.Logger
		c.Next() // executes after next chained func finishes its work

		latency := time.Since(currentTime)
		logger.Error("🌟🌟REQUEST INFORMATION 🌟🌟",
			// * stored in key val pairs
			slog.String("method",c.Request.Method),
			slog.String("path",reqPath),
			slog.String("query",queryParam),
			// ! when there is to convert rune to string -> use Sprintf - to format into string
			slog.String("latency",fmt.Sprintf(":%c",latency)), //! c placeholder fpr rune 
			slog.String("client-ip",c.ClientIP()),
			// slog.String("status-code",fmt.Sprintf(":%c",c.Writer.Status())),
			// slog.String("body_size",fmt.Sprintf(":%c",c.Writer.Size())),

		)


		if len(c.Errors) > 0 {
			for _,err := range c.Errors {
				slog.Error("request errors",slog.String("error",err.Error()))
			}
		}

	}
}