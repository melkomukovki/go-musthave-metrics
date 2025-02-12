package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
)

func SubnetValidatorMiddleware(subnet string) gin.HandlerFunc {
	_, netCidr, err := net.ParseCIDR(subnet)
	if err != nil {
		log.Fatal().Err(err).Str("subnet", subnet).Msg("invalid subnet")
	}

	return func(c *gin.Context) {
		fromRequest := c.GetHeader("X-Real-IP")
		log.Debug().Str("cidr", netCidr.String()).Str("clientIPHeader", fromRequest).Msg("subnet validator")

		if fromRequest == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "X-Real-IP header is empty"})
			return
		}

		ip := net.ParseIP(fromRequest)
		if ip == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Invalid IP address"})
			return
		}

		if !netCidr.Contains(ip) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "IP address not allowed"})
			return
		}
		c.Next()
	}
}
