package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MemStorage struct {
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
}

func (m MemStorage) AddGaugeMetric(name, v string) error {
	vFloat, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return errors.New("can't parse value")
	}
	m.GaugeMetrics[name] = vFloat
	return nil
}

func (m MemStorage) AddCounterMetric(name, v string) error {
	vInt, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return errors.New("can't parse value")
	}
	if val, ok := m.CounterMetrics[name]; ok {
		newVal := val + vInt
		m.CounterMetrics[name] = newVal
	} else {
		m.CounterMetrics[name] = vInt
	}
	return nil
}

func (m MemStorage) GetGaugeMetric(name string) (float64, error) {
	v, ok := m.GaugeMetrics[name]
	if ok {
		return v, nil
	}
	return 0, errors.New("value not found")
}

func (m MemStorage) GetCounterMetric(name string) (int64, error) {
	v, ok := m.CounterMetrics[name]
	if ok {
		return v, nil
	}
	return 0, errors.New("value not found")
}

func (m MemStorage) GetAllMetrics() string {
	res := ""
	for k, v := range m.GaugeMetrics {
		res += fmt.Sprintf("%s:%.6f\n", k, v)
	}
	for k, v := range m.CounterMetrics {
		res += fmt.Sprintf("%s:%d\n", k, v)
	}
	return res
}

type Storage interface {
	AddGaugeMetric(string, string) error
	AddCounterMetric(string, string) error
	GetGaugeMetric(string) (float64, error)
	GetCounterMetric(string) (int64, error)
	GetAllMetrics() string
}

var storage Storage = MemStorage{
	GaugeMetrics:   make(map[string]float64),
	CounterMetrics: make(map[string]int64),
}

func PostMetricHandler(c *gin.Context) {
	// if !slices.Contains(c.Request.Header["Content-Type"], "text/plain") {
	// 	c.String(http.StatusBadRequest, `No header "text/plain"`)
	// 	return
	// }

	mType := c.Params.ByName("mType")
	mName := c.Params.ByName("mName")
	mValue := c.Params.ByName("mValue")

	if mType == "gauge" {
		err := storage.AddGaugeMetric(mName, mValue)
		if err != nil {
			c.String(http.StatusBadRequest, "Can't add gauge metric: %s - %s", mName, mValue)
			return
		}
	} else if mType == "counter" {
		err := storage.AddCounterMetric(mName, mValue)
		if err != nil {
			c.String(http.StatusBadRequest, "Can't add counter metric: %s - %s", mName, mValue)
			return
		}
	} else {
		c.String(http.StatusBadRequest, "Invalid metric type: %s", mType)
		return
	}
	c.String(http.StatusOK, "OK")
}

func GetMetricHandler(c *gin.Context) {
	mType := c.Params.ByName("mType")
	mName := c.Params.ByName("mName")

	if mType == "gauge" {
		mV, err := storage.GetGaugeMetric(mName)
		if err != nil {
			c.String(http.StatusNotFound, "Can't found metric")
			return
		}
		c.String(http.StatusOK, "%.6f", mV)
		return
	} else if mType == "counter" {
		mV, err := storage.GetCounterMetric(mName)
		if err != nil {
			c.String(http.StatusNotFound, "Can't found metric")
			return
		}
		c.String(http.StatusOK, "%d", mV)
		return
	}
	c.String(http.StatusBadRequest, "Invalid metric type")
}

func ShowMetrics(c *gin.Context) {
	res := storage.GetAllMetrics()
	c.String(http.StatusOK, res)
}

func main() {
	gin.ForceConsoleColor()
	r := gin.Default()

	r.POST("/update/:mType/:mName/:mValue", PostMetricHandler)
	r.GET("/value/:mType/:mName", GetMetricHandler)
	r.GET("/", ShowMetrics)

	r.Run(":8080")
}
