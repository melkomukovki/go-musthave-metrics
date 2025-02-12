// Package middleware provides functions used when processing requests
package middleware

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"github.com/gin-gonic/gin"
	pc "github.com/melkomukovki/go-musthave-metrics/internal/crypto"
	"io"
	"net/http"
)

func CryptoMiddleware(key *rsa.PrivateKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost {

			encBody, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": err.Error(),
				})
				c.Abort()
				return
			}

			fmt.Println("before decrypt:", string(encBody))
			decryptedBody, err := pc.Decrypt(encBody, key)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				c.Abort()
				return
			}

			c.Request.Body = io.NopCloser(bytes.NewReader(decryptedBody))
		}
		c.Next()
	}
}
