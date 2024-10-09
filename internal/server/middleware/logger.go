package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		timeStamp := time.Now()
		latencyRaw := timeStamp.Sub(start).Seconds()
		latency := fmt.Sprintf("%.3fs", latencyRaw)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		log.Info().
			Str("clientIP", clientIP).
			Str("method", method).
			Str("URL", path).
			Int("statusCode", statusCode).
			Str("requestTime", latency).
			Int("bodySizeBytes", bodySize).
			Msg("request recived")
	}

}
