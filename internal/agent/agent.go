package agent

import (
	"fmt"
	"math/rand/v2"
	"runtime"

	"github.com/go-resty/resty/v2"
	"github.com/melkomukovki/go-musthave-metrics/internal/config"
	"github.com/melkomukovki/go-musthave-metrics/internal/metrics"
)

type Agent struct {
	pollCounter    int64
	counterMetrics []metrics.CounterMetric
	gaugeMetrics   []metrics.GaugeMetric
	config         *config.ClientConfig
}

func NewAgent(cfg *config.ClientConfig) *Agent {
	return &Agent{
		config: cfg,
	}
}

func (a *Agent) addPollCounter() {
	a.pollCounter++
}

func (a *Agent) GetPollInterval() int {
	return a.config.PollInterval
}

func (a *Agent) GetReportInterval() int {
	return a.config.ReportInterval
}

func (a *Agent) PollMetrics() {
	a.addPollCounter()
	a.getGaugeMetrics()
	a.getCounterMetrics()
}

func (a *Agent) ReportMetrics(c *resty.Client) {
	for _, v := range a.gaugeMetrics {
		sC, err := a.sendMetric(c, "gauge", v.Name, v.Value)
		if err != nil {
			fmt.Printf("Detected error: %s\n", err)
			continue
		}
		fmt.Printf("Metric %s Value %v sended. Status code: %d\n", v.Name, v.Value, sC)
	}
	for _, v := range a.counterMetrics {
		sC, err := a.sendMetric(c, "counter", v.Name, v.Value)
		if err != nil {
			fmt.Printf("Detected error: %s\n", err)
			continue
		}
		fmt.Printf("Metric %s Value %v sended. Status code: %d\n", v.Name, v.Value, sC)
	}
}

func (a *Agent) getGaugeMetrics() {
	var newAr []metrics.GaugeMetric
	var rtm runtime.MemStats

	runtime.ReadMemStats(&rtm)

	newAr = append(newAr, metrics.GaugeMetric{Name: "Alloc", Value: float64(rtm.Alloc)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "BuckHashSys", Value: float64(rtm.BuckHashSys)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "Frees", Value: float64(rtm.Frees)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "GCCPUFraction", Value: float64(rtm.GCCPUFraction)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "GCSys", Value: float64(rtm.GCSys)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "HeapAlloc", Value: float64(rtm.HeapAlloc)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "HeapIdle", Value: float64(rtm.HeapIdle)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "HeapInuse", Value: float64(rtm.HeapInuse)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "HeapObjects", Value: float64(rtm.HeapObjects)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "HeapReleased", Value: float64(rtm.HeapReleased)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "HeapSys", Value: float64(rtm.HeapSys)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "LastGC", Value: float64(rtm.LastGC)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "Lookups", Value: float64(rtm.Lookups)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "MCacheInuse", Value: float64(rtm.MCacheInuse)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "MCacheSys", Value: float64(rtm.MCacheSys)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "MSpanInuse", Value: float64(rtm.MSpanInuse)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "MSpanSys", Value: float64(rtm.MSpanSys)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "Mallocs", Value: float64(rtm.Mallocs)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "NextGC", Value: float64(rtm.NextGC)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "NumForcedGC", Value: float64(rtm.NumForcedGC)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "NumGC", Value: float64(rtm.NumGC)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "OtherSys", Value: float64(rtm.OtherSys)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "PauseTotalNs", Value: float64(rtm.PauseTotalNs)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "StackInuse", Value: float64(rtm.StackInuse)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "StackSys", Value: float64(rtm.StackSys)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "Sys", Value: float64(rtm.Sys)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "TotalAlloc", Value: float64(rtm.TotalAlloc)})
	newAr = append(newAr, metrics.GaugeMetric{Name: "RandomValue", Value: rand.Float64()})

	a.gaugeMetrics = newAr
}

func (a *Agent) getCounterMetrics() {
	var newAr []metrics.CounterMetric
	newAr = append(newAr, metrics.CounterMetric{Name: "PollCount", Value: a.pollCounter})
	a.counterMetrics = newAr
}

func (a *Agent) sendMetric(c *resty.Client, mType string, mName string, mValue interface{}) (int, error) {
	fullURL := fmt.Sprintf("http://%s/update/%s/%s/%v", a.config.Address, mType, mName, mValue)
	resp, err := c.R().
		SetHeader("Content-Type", "text/plain").
		Post(fullURL)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode(), nil
}
