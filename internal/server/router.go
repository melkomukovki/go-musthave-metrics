package server

import (
	"github.com/gin-gonic/gin"
	"github.com/melkomukovki/go-musthave-metrics/internal/handlers"
	"github.com/melkomukovki/go-musthave-metrics/internal/storage"
)

func NewServerRouter(store storage.Storage) *gin.Engine {
	r := gin.Default()
	r.POST("/update/:mType/:mName/:mValue", handlers.PostMetricHandler(store))
	r.GET("/value/:mType/:mName", handlers.GetMetricHandler(store))
	r.GET("/", handlers.ShowMetrics(store))

	return r
}
