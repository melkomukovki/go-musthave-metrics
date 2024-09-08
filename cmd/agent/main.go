package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/melkomukovki/go-musthave-metrics/internal/config"
)

type GaugeMetric struct {
	Name  string
	Value float64
}

type CounterMetric struct {
	Name  string
	Value int64
}

var (
	pollCounter    int64 = 0
	counterMetrics []CounterMetric
	gaugeMetrics   []GaugeMetric
)

func getGaugeMetrics(ar *[]GaugeMetric) {
	var newAr []GaugeMetric
	var rtm runtime.MemStats

	runtime.ReadMemStats(&rtm)

	// Collect metrics
	newAr = append(newAr, GaugeMetric{Name: "Alloc", Value: float64(rtm.Alloc)})
	newAr = append(newAr, GaugeMetric{Name: "BuckHashSys", Value: float64(rtm.BuckHashSys)})
	newAr = append(newAr, GaugeMetric{Name: "Frees", Value: float64(rtm.Frees)})
	newAr = append(newAr, GaugeMetric{Name: "GCCPUFraction", Value: float64(rtm.GCCPUFraction)})
	newAr = append(newAr, GaugeMetric{Name: "GCSys", Value: float64(rtm.GCSys)})
	newAr = append(newAr, GaugeMetric{Name: "HeapAlloc", Value: float64(rtm.HeapAlloc)})
	newAr = append(newAr, GaugeMetric{Name: "HeapIdle", Value: float64(rtm.HeapIdle)})
	newAr = append(newAr, GaugeMetric{Name: "HeapInuse", Value: float64(rtm.HeapInuse)})
	newAr = append(newAr, GaugeMetric{Name: "HeapObjects", Value: float64(rtm.HeapObjects)})
	newAr = append(newAr, GaugeMetric{Name: "HeapReleased", Value: float64(rtm.HeapReleased)})
	newAr = append(newAr, GaugeMetric{Name: "HeapSys", Value: float64(rtm.HeapSys)})
	newAr = append(newAr, GaugeMetric{Name: "LastGC", Value: float64(rtm.LastGC)})
	newAr = append(newAr, GaugeMetric{Name: "Lookups", Value: float64(rtm.Lookups)})
	newAr = append(newAr, GaugeMetric{Name: "MCacheInuse", Value: float64(rtm.MCacheInuse)})
	newAr = append(newAr, GaugeMetric{Name: "MCacheSys", Value: float64(rtm.MCacheSys)})
	newAr = append(newAr, GaugeMetric{Name: "MSpanInuse", Value: float64(rtm.MSpanInuse)})
	newAr = append(newAr, GaugeMetric{Name: "MSpanSys", Value: float64(rtm.MSpanSys)})
	newAr = append(newAr, GaugeMetric{Name: "Mallocs", Value: float64(rtm.Mallocs)})
	newAr = append(newAr, GaugeMetric{Name: "NextGC", Value: float64(rtm.NextGC)})
	newAr = append(newAr, GaugeMetric{Name: "NumForcedGC", Value: float64(rtm.NumForcedGC)})
	newAr = append(newAr, GaugeMetric{Name: "NumGC", Value: float64(rtm.NumGC)})
	newAr = append(newAr, GaugeMetric{Name: "OtherSys", Value: float64(rtm.OtherSys)})
	newAr = append(newAr, GaugeMetric{Name: "PauseTotalNs", Value: float64(rtm.PauseTotalNs)})
	newAr = append(newAr, GaugeMetric{Name: "StackInuse", Value: float64(rtm.StackInuse)})
	newAr = append(newAr, GaugeMetric{Name: "StackSys", Value: float64(rtm.StackSys)})
	newAr = append(newAr, GaugeMetric{Name: "Sys", Value: float64(rtm.Sys)})
	newAr = append(newAr, GaugeMetric{Name: "TotalAlloc", Value: float64(rtm.TotalAlloc)})
	newAr = append(newAr, GaugeMetric{Name: "RandomValue", Value: rand.Float64()})

	*ar = newAr
}

func getCounterMetrics(ar *[]CounterMetric) {
	var newAr []CounterMetric
	newAr = append(newAr, CounterMetric{Name: "PollCount", Value: pollCounter})
	*ar = newAr
}

func updateMetrics(pollInterval int) {
	for {
		pollCounter++
		getCounterMetrics(&counterMetrics)
		getGaugeMetrics(&gaugeMetrics)
		time.Sleep(time.Second * time.Duration(pollInterval))
	}
}

func sendMetrics(c *resty.Client, url string, reportInterval int) {
	url = "http://" + url
	for {
		for _, v := range gaugeMetrics {
			url := fmt.Sprintf("%s/update/gauge/%s/%.6f", url, v.Name, v.Value)
			resp, err := c.R().
				SetHeader("Content-Type", "text/plain").
				Post(url)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(resp.StatusCode())
		}

		for _, v := range counterMetrics {
			url := fmt.Sprintf("%s/update/counter/%s/%d", url, v.Name, v.Value)
			resp, err := c.R().
				SetHeader("Content-Type", "text/plain").
				Post(url)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(resp.StatusCode())
		}

		time.Sleep(time.Second * time.Duration(reportInterval))
	}
}

func main() {
	cfg := config.GetClientConfig()
	fmt.Println(cfg)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	client := resty.New()

	go updateMetrics(cfg.PollInterval)
	go sendMetrics(client, cfg.Address, cfg.ReportInterval)

	<-sigs
}
