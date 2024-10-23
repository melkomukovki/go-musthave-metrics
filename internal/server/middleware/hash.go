package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type bodyWriter struct {
	gin.ResponseWriter
	body    *bytes.Buffer
	hashKey string
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		hash := getHash(w.body.Bytes(), w.hashKey)
		w.ResponseWriter.Header().Set("HashSHA256", hash)
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func HashSumMiddleware(hashKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost {

			rawBody, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": err.Error(),
				})
				c.Abort()
				return
			}

			if !validateRecivedHash(rawBody, c.GetHeader("HashSHA256"), hashKey) {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "invalid hash value",
				})
				c.Abort()
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(rawBody))
		}

		bw := &bodyWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
			hashKey:        hashKey,
		}
		c.Writer = bw

		c.Next()

		// respBody := bw.body.Bytes()
		// respHash := getHash(respBody, hashKey)
		// respHashStr := hex.EncodeToString(respHash)

		// c.Header("HashSHA256", respHashStr)

	}
}

func validateRecivedHash(data []byte, hash, hashKey string) bool {
	h := hmac.New(sha256.New, []byte(hashKey))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)) == hash
}

func getHash(data []byte, hashKey string) string {
	h := hmac.New(sha256.New, []byte(hashKey))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
