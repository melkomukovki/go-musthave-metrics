// Package middleware provides functions used when processing requests
package middleware

import (
	"compress/gzip"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type compressWriter struct {
	gin.ResponseWriter
	zw *gzip.Writer
}

func (c *compressWriter) Write(data []byte) (int, error) {

	return c.zw.Write(data)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	c.ResponseWriter.WriteHeader(statusCode)
}

// GzipMiddleware adds support compression for request and response
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": "invalid gzip format",
				})
				return
			}
			defer func() {
				if err := gzipReader.Close(); err != nil {
					log.Error().Err(err).Msg("failed to close gzip reader")
				}
			}()
			c.Request.Body = io.NopCloser(gzipReader)
		}

		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			gzipWriter := gzip.NewWriter(c.Writer)
			defer func() {
				if err := gzipWriter.Close(); err != nil {
					log.Error().Err(err).Msg("failed to close gzip writer")
				}
			}()
			c.Writer = &compressWriter{ResponseWriter: c.Writer, zw: gzipWriter}
		}
		c.Next()
	}
}
