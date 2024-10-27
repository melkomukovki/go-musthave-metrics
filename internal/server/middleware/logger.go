package middleware

import (
	"bytes"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		timeStamp := time.Now()
		latencyRaw := timeStamp.Sub(start).Seconds()
		latency := fmt.Sprintf("%.3fs", latencyRaw)
		clientIP := c.ClientIP()
		method := c.Request.Method
		hashInHeader := c.Request.Header.Get("HashSHA256")
		hashOutHeader := c.Writer.Header().Get("HashSHA256")
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()
		respBody := blw.body.String()

		log.Info().
			Str("clientIP", clientIP).
			Str("method", method).
			Str("URL", path).
			Int("statusCode", statusCode).
			Str("requestTime", latency).
			Int("bodySizeBytes", bodySize).
			Str("hashIn", hashInHeader).
			Str("hashOut", hashOutHeader).
			Str("responseBody", respBody).
			Msg("request recived")
	}

}
