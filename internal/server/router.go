package server

import (
	"github.com/gin-gonic/gin"
	"github.com/melkomukovki/go-musthave-metrics/internal/server/handlers"
	"github.com/melkomukovki/go-musthave-metrics/internal/server/middleware"
	"github.com/melkomukovki/go-musthave-metrics/internal/storage"
)

func NewServerRouter(store storage.Storage) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(middleware.LoggerMiddleware(), gin.Recovery())
	engine.Use(middleware.GzipMiddleware(), gin.Recovery())

	engine.POST("/update/", handlers.PostMetricHandlerJSON(store))
	engine.POST("/update/:mType/:mName/:mValue", handlers.PostMetricHandler(store))

	engine.POST("/value/", handlers.GetMetricHandlerJSON(store))
	engine.GET("/value/:mType/:mName", handlers.GetMetricHandler(store))

	engine.GET("/ping", handlers.PingHandler(store))

	engine.GET("/", handlers.ShowMetrics(store))

	return engine
}
