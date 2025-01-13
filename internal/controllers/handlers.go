// Package controllers describe handlers used in project
package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/melkomukovki/go-musthave-metrics/internal/controllers/middleware"
	"github.com/melkomukovki/go-musthave-metrics/internal/entities"
	"github.com/melkomukovki/go-musthave-metrics/internal/services"
)

// AppHandler define handler structure
type AppHandler struct {
	Service *services.Service
}

// NewHandler adds needed routers and middleware to our gin engine
func NewHandler(router *gin.Engine, service *services.Service, hashKey string) {
	handler := AppHandler{Service: service}

	appRoutes := router.Group("/")
	appRoutes.Use(middleware.LoggerMiddleware(), gin.Recovery())
	appRoutes.Use(middleware.GzipMiddleware())
	if hashKey != "" {
		appRoutes.Use(middleware.HashSumMiddleware(hashKey))
	}
	{
		appRoutes.POST("/update/", handler.postMetricJSON)
		appRoutes.POST("/updates/", handler.postMultipleMetrics)
		appRoutes.POST("/update/:mType/:mName/mValue", handler.postMetric)

		appRoutes.POST("/value/", handler.getMetricJSON)
		appRoutes.GET("/value/:mType/:mName", handler.getMetric)

		appRoutes.GET("/ping", handler.ping)

		appRoutes.GET("/", handler.showMetrics)
	}
}

func (a *AppHandler) postMetricJSON(c *gin.Context) {
	var v entities.Metric
	if err := c.BindJSON(&v); err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			gin.H{"message": fmt.Sprintf("Invalid payload. Error: %s", err.Error())},
		)
		return
	}

	err := a.Service.AddMetric(c, v)
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			gin.H{
				"message": fmt.Sprintf("Invalid payload. Error: %s", err.Error()), // TODO
			})
		return
	}

	rM, err := a.Service.GetMetric(c, v.MType, v.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, rM)
}

func (a *AppHandler) postMultipleMetrics(c *gin.Context) {
	var metrics []entities.Metric
	if err := c.BindJSON(&metrics); err != nil {
		errMsg := fmt.Sprintf("Invalid payload. Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"message": errMsg,
		})
		return
	}

	err := a.Service.AddMultipleMetrics(c, metrics)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// TODO
func (a *AppHandler) postMetric(c *gin.Context) {
	mType := c.Params.ByName("mType")
	mName := c.Params.ByName("mName")
	mValue := c.Params.ByName("mValue")

	switch mType {
	case entities.Gauge:
		value, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "Can't convert value %s to float64", mValue)
		}
		metric := entities.Metric{ID: mName, MType: mType, Value: &value}
		err = a.Service.AddMetric(c, metric)
		if err != nil {
			c.String(http.StatusBadRequest, "Can't add gauge metric: %s - %s. Error: %s", mName, mValue, err.Error())
			return
		}
	case entities.Counter:
		value, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "Can't convert value %s to int64", mValue)
		}
		metric := entities.Metric{ID: mName, MType: mType, Delta: &value}
		err = a.Service.AddMetric(c, metric)
		if err != nil {
			c.String(http.StatusBadRequest, "Can't add counter metric: %s - %s. Error: %s", mName, mValue, err.Error())
			return
		}
	default:
		c.String(http.StatusBadRequest, "Invalid metric type: %s", mType)
		return
	}

	c.String(http.StatusOK, "OK")
}

func (a *AppHandler) getMetricJSON(c *gin.Context) {
	var v entities.Metric
	if err := c.BindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid payload " + err.Error(),
		})
		return
	}

	res, err := a.Service.GetMetric(c, v.MType, v.ID)
	if err != nil {
		if errors.Is(err, entities.ErrMetricNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, res)
}

func (a *AppHandler) getMetric(c *gin.Context) {
	mType := c.Params.ByName("mType")
	mName := c.Params.ByName("mName")

	metric, err := a.Service.GetMetric(c, mType, mName)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}

	switch metric.MType {
	case entities.Gauge:
		fV := fmt.Sprintf("%.3f", *metric.Value)
		fV = strings.TrimRight(strings.TrimRight(fV, "0"), ".")
		c.String(http.StatusOK, fV)
		return
	case entities.Counter:
		c.String(http.StatusOK, "%d", *metric.Delta)
		return
	default:
		c.String(http.StatusInternalServerError, "unexpected metric type from store")
	}
}

func (a *AppHandler) ping(c *gin.Context) {
	err := a.Service.Ping(c)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
	} else {
		c.String(http.StatusOK, "Pong")
	}
}

func (a *AppHandler) showMetrics(c *gin.Context) {
	metrics, err := a.Service.GetAllMetrics(c)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}

	var result string
	for _, v := range metrics {
		switch v.MType {
		case entities.Gauge:
			result += fmt.Sprintf("%s:%.3f\n", v.ID, *v.Value)
		case entities.Counter:
			result += fmt.Sprintf("%s:%d\n", v.ID, *v.Delta)
		}
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(result))
}
