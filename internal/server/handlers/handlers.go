package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/melkomukovki/go-musthave-metrics/internal/storage"
)

func PostMetricHandlerJSON(store storage.Storage) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var v storage.Metrics
		if err := c.BindJSON(&v); err != nil {
			errMsg := fmt.Sprintf("Invalid payload. Error: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"message": errMsg,
			})
			return
		}

		err := store.AddMetric(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		rM, err := store.GetMetric(v.MType, v.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, rM)
	}
	return fn
}

func PostMetricHandler(store storage.Storage) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		mType := c.Params.ByName("mType")
		mName := c.Params.ByName("mName")
		mValue := c.Params.ByName("mValue")

		switch mType {
		case storage.Gauge:
			value, err := strconv.ParseFloat(mValue, 64)
			if err != nil {
				c.String(http.StatusBadRequest, "Can't convert value %s to float64", mValue)
			}
			metric := storage.Metrics{ID: mName, MType: mType, Value: &value}
			err = store.AddMetric(metric)
			if err != nil {
				c.String(http.StatusBadRequest, "Can't add gauge metric: %s - %s. Error: %s", mName, mValue, err.Error())
				return
			}
		case storage.Counter:
			value, err := strconv.ParseInt(mValue, 10, 64)
			if err != nil {
				c.String(http.StatusBadRequest, "Can't convert value %s to int64", mValue)
			}
			metric := storage.Metrics{ID: mName, MType: mType, Delta: &value}
			err = store.AddMetric(metric)
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
	return fn
}

func GetMetricHandlerJSON(store storage.Storage) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var v storage.Metrics
		if err := c.BindJSON(&v); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invaild payload " + err.Error(),
			})
			return
		}

		res, err := store.GetMetric(v.MType, v.ID)
		if err != nil {
			if err.Error() == "metric not found" {
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
	return fn
}

func GetMetricHandler(store storage.Storage) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		mType := c.Params.ByName("mType")
		mName := c.Params.ByName("mName")

		metric, err := store.GetMetric(mType, mName)
		if err != nil {
			c.String(http.StatusNotFound, err.Error())
			return
		}

		switch metric.MType {
		case storage.Gauge:
			fV := fmt.Sprintf("%.3f", *metric.Value)
			fV = strings.TrimRight(strings.TrimRight(fV, "0"), ".")
			c.String(http.StatusOK, fV)
			return
		case storage.Counter:
			c.String(http.StatusOK, "%d", *metric.Delta)
			return
		default:
			c.String(http.StatusInternalServerError, "unexpected metric type from store")
		}
	}
	return fn
}

func ShowMetrics(store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var result string
		metrics := store.GetAllMetrics()
		for _, v := range metrics {
			switch v.MType {
			case storage.Gauge:
				result += fmt.Sprintf("%s:%.3f\n", v.ID, *v.Value)
			case storage.Counter:
				result += fmt.Sprintf("%s:%d\n", v.ID, *v.Delta)
			}
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(result))
	}
}
