package middleware

import (
	"compress/gzip"
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

func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		contentEncoding := c.GetHeader("Content-Encoding")
		recivedGzip := strings.Contains(contentEncoding, "gzip")

		if recivedGzip {
			gzipReader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "invalid gzip format",
				})
				c.Abort()
				return
			}
			defer gzipReader.Close()
			c.Request.Body = io.NopCloser(gzipReader)
		}

		acceptEncoding := c.GetHeader("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip {
			cw := &compressWriter{
				ResponseWriter: c.Writer,
				zw:             gzip.NewWriter(c.Writer),
			}
			c.Writer = cw
			defer cw.zw.Close()
		}
		c.Next()
	}
}
