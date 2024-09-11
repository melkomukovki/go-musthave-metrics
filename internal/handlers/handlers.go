package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/melkomukovki/go-musthave-metrics/internal/storage"
)

func PostMetricHandler(store storage.Storage) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		mType := c.Params.ByName("mType")
		mName := c.Params.ByName("mName")
		mValue := c.Params.ByName("mValue")

		switch mType {
		case "gauge":
			err := store.AddGaugeMetric(mName, mValue)
			if err != nil {
				c.String(http.StatusBadRequest, "Can't add gauge metric: %s - %s", mName, mValue)
				return
			}
		case "counter":
			err := store.AddCounterMetric(mName, mValue)
			if err != nil {
				c.String(http.StatusBadRequest, "Can't add counter metric: %s - %s", mName, mValue)
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

func GetMetricHandler(store storage.Storage) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		mType := c.Params.ByName("mType")
		mName := c.Params.ByName("mName")

		if mType == "gauge" {
			mV, err := store.GetGaugeMetric(mName)
			if err != nil {
				c.String(http.StatusNotFound, "Can't found metric")
				return
			}
			fV := fmt.Sprintf("%.3f", mV)
			fV = strings.TrimRight(strings.TrimRight(fV, "0"), ".")
			c.String(http.StatusOK, fV)
			return
		} else if mType == "counter" {
			mV, err := store.GetCounterMetric(mName)
			if err != nil {
				c.String(http.StatusNotFound, "Can't found metric")
				return
			}
			c.String(http.StatusOK, "%d", mV)
			return
		}
		c.String(http.StatusBadRequest, "Invalid metric type")
	}
	return fn
}

func ShowMetrics(store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		res := store.GetAllMetrics()
		c.String(http.StatusOK, res)
	}
}
