package server

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/melkomukovki/go-musthave-metrics/internal/handlers"
	"github.com/melkomukovki/go-musthave-metrics/internal/storage"
	"github.com/sirupsen/logrus"
)

const timeFormat = "02/Jan/2006 15:04:05 -0700"

func LoggerMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		timeStamp := time.Now()
		timeStampFormated := timeStamp.Format(timeFormat)
		latency := timeStamp.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		log.WithFields(logrus.Fields{
			"timeStamp":   timeStampFormated,
			"clientIP":    clientIP,
			"method":      method,
			"URL":         path,
			"statusCode":  statusCode,
			"requestTime": latency,
			"bodySize":    bodySize,
		}).Info("request recived")
	}

}

func NewServerRouter(store storage.Storage) *gin.Engine {
	// Disable debug info
	// gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// Custom log
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	r.Use(LoggerMiddleware(log), gin.Recovery())

	r.POST("/update/", handlers.PostMetricHandlerJSON(store))
	r.POST("/update/:mType/:mName/:mValue", handlers.PostMetricHandler(store))
	r.POST("/value/", handlers.GetMetricHandlerJSON(store))
	r.GET("/value/:mType/:mName", handlers.GetMetricHandler(store))
	r.GET("/", handlers.ShowMetrics(store))

	return r
}
