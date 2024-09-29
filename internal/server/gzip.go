package server

import (
	"bufio"
	"compress/gzip"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type compressWriter struct {
	w  gin.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(r gin.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  r,
		zw: gzip.NewWriter(r),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(data []byte) (int, error) {
	return c.zw.Write(data)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) CloseNotify() <-chan bool {
	return c.w.CloseNotify()
}

func (c *compressWriter) Flush() {
	c.w.Flush()
}

func (c *compressWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return c.w.Hijack()
}

func (c *compressWriter) Pusher() http.Pusher {
	return c.w.Pusher()
}

func (c *compressWriter) Size() int {
	return c.w.Size()
}

func (c *compressWriter) Status() int {
	return c.w.Status()
}

func (c *compressWriter) WriteHeaderNow() {
	c.w.WriteHeaderNow()
}

func (c *compressWriter) WriteString(s string) (int, error) {
	return c.w.WriteString(s)
}

func (c *compressWriter) Written() bool {
	return c.w.Written()
}

func gzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		applicationHeader := c.GetHeader("Content-Type")
		if strings.Contains(applicationHeader, "application/json") || strings.Contains(applicationHeader, "text/html") {
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
				c.Request.Body = gzipReader
			}

			acceptEncoding := c.GetHeader("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")

			if supportsGzip {
				cw := newCompressWriter(c.Writer)
				c.Writer = cw
				defer cw.zw.Close()
			}
		}
		c.Next()
	}
}
