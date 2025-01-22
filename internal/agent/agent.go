// Package agent определяет структуру и функции агента по отправке метрик
package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/melkomukovki/go-musthave-metrics/internal/utils"
	"github.com/rs/zerolog/log"
	"math/rand/v2"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/melkomukovki/go-musthave-metrics/internal/entities"

	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/melkomukovki/go-musthave-metrics/internal/agent/config"
)

// Agent описывает структуру агента по сбору и отправке метрик
type Agent struct {
	mu              sync.Mutex
	pollCounter     int64
	metrics         []entities.Metric
	config          *config.ClientConfig
	workerPool      chan func()
	cryptoPublicKey *rsa.PublicKey
}

// NewAgent функция для получения экземпляра агента
func NewAgent(cfg *config.ClientConfig) (*Agent, error) {
	// init base struct
	agent := &Agent{
		config:     cfg,
		workerPool: make(chan func(), cfg.RateLimit),
	}

	// if cert path provided, get cert
	if cfg.CryptoKey != "" {
		publicKey, err := utils.GetPublicKey(cfg.CryptoKey)
		if err != nil {
			return nil, err
		}
		agent.cryptoPublicKey = publicKey
	}

	agent.createWorkerPool()

	return agent, nil
}

func (a *Agent) addPollCounter() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pollCounter++
}

func (a *Agent) resetPollCounter() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pollCounter = 0
}

func createGaugeMetric(id string, value float64, mType string) entities.Metric {
	return entities.Metric{ID: id, MType: mType, Value: &value}
}

func (a *Agent) pollAdditionalMetrics() []entities.Metric {
	var sMetrics []entities.Metric

	v, err := mem.VirtualMemory()
	if err != nil {
		log.Error().Err(err).Msg("Error getting memory metrics")
		return sMetrics
	}

	metricsData := []struct {
		id    string
		value float64
	}{
		{"TotalMemory", float64(v.Total)},
		{"FreeMemory", float64(v.Free)},
	}

	for _, m := range metricsData {
		sMetrics = append(sMetrics, createGaugeMetric(m.id, m.value, entities.Gauge))
	}

	cpuPercentages, err := cpu.Percent(0, true)
	if err != nil {
		fmt.Printf("Error collecting CPU metrics: %s\n", err.Error())
	}

	for i, m := range cpuPercentages {
		id := fmt.Sprintf("CPUutilization%d", i+1)
		sMetrics = append(sMetrics, createGaugeMetric(id, m, entities.Gauge))
	}

	return sMetrics
}

func (a *Agent) pollMetrics() {
	a.addPollCounter()

	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	metricsData := []struct {
		id    string
		value float64
	}{
		{"Alloc", float64(rtm.Alloc)},
		{"BuckHashSys", float64(rtm.BuckHashSys)},
		{"Frees", float64(rtm.Frees)},
		{"GCCPUFraction", rtm.GCCPUFraction},
		{"GCSys", float64(rtm.GCSys)},
		{"HeapAlloc", float64(rtm.HeapAlloc)},
		{"HeapIdle", float64(rtm.HeapIdle)},
		{"HeapInuse", float64(rtm.HeapInuse)},
		{"HeapObjects", float64(rtm.HeapObjects)},
		{"HeapReleased", float64(rtm.HeapReleased)},
		{"HeapSys", float64(rtm.HeapSys)},
		{"LastGC", float64(rtm.LastGC)},
		{"Lookups", float64(rtm.Lookups)},
		{"MCacheInuse", float64(rtm.MCacheInuse)},
		{"MCacheSys", float64(rtm.MCacheSys)},
		{"MSpanInuse", float64(rtm.MSpanInuse)},
		{"MSpanSys", float64(rtm.MSpanSys)},
		{"Mallocs", float64(rtm.Mallocs)},
		{"NextGC", float64(rtm.NextGC)},
		{"NumForcedGC", float64(rtm.NumForcedGC)},
		{"NumGC", float64(rtm.NumGC)},
		{"OtherSys", float64(rtm.OtherSys)},
		{"PauseTotalNs", float64(rtm.PauseTotalNs)},
		{"StackInuse", float64(rtm.StackInuse)},
		{"StackSys", float64(rtm.StackSys)},
		{"Sys", float64(rtm.Sys)},
		{"TotalAlloc", float64(rtm.TotalAlloc)},
		{"RandomValue", rand.Float64()},
	}

	var newAr []entities.Metric
	for _, m := range metricsData {
		newAr = append(newAr, createGaugeMetric(m.id, m.value, entities.Gauge))
	}

	newAr = append(newAr, entities.Metric{ID: "PollCount", MType: entities.Counter, Delta: &a.pollCounter})

	newAr = append(newAr, a.pollAdditionalMetrics()...)

	a.mu.Lock()
	defer a.mu.Unlock()
	a.metrics = newAr
}

func (a *Agent) getHash(data []byte) string {
	h := hmac.New(sha256.New, []byte(a.config.HashKey))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func (a *Agent) createWorkerPool() {
	for i := 0; i < a.config.RateLimit; i++ {
		go func() {
			for task := range a.workerPool {
				task()
			}
		}()
	}
}

func (a *Agent) reportMetrics(c *resty.Client) {
	a.workerPool <- func() {
		url := fmt.Sprintf("http://%s/updates/", a.config.Address)

		headers := map[string]string{
			"Content-Type":     "application/json",
			"Content-Encoding": "gzip",
			"Accept-Encoding":  "gzip",
		}

		mMarshaled, err := json.Marshal(a.metrics)
		if err != nil {
			log.Error().Err(err).Msg("Error marshalling metrics")
			return
		}
		if a.config.HashKey != "" {
			hashString := a.getHash(mMarshaled)
			headers["HashSHA256"] = hashString
		}

		// encrypt message if certificate presents
		if a.config.CryptoKey != "" {
			mMarshaled, err = utils.Encrypt(mMarshaled, a.cryptoPublicKey)
			if err != nil {
				log.Error().Err(err).Msg("Error encrypting metrics")
				return
			}
		}

		compressedData, err := gzipData(mMarshaled)
		if err != nil {
			log.Error().Err(err).Msg("Error compressing metrics")
			return
		}

		resp, err := c.R().
			SetBody(compressedData).
			SetHeaders(headers).
			Post(url)
		if err != nil {
			log.Error().Err(err).Msg("Error reporting metrics")
			return
		}
		log.Info().Msg("Metric was sent")
		fmt.Println(resp.Header())
		a.resetPollCounter()
	}
}

// Run - функция для запуска работы агента
func (a *Agent) Run(cResty *resty.Client) {
	pollTicker := time.NewTicker(time.Duration(a.config.PollInterval) * time.Second)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(time.Duration(a.config.ReportInterval) * time.Second)
	defer reportTicker.Stop()

	sigsEnd := make(chan os.Signal, 1)
	signal.Notify(sigsEnd, syscall.SIGINT, syscall.SIGTERM)

	// First-time poll and send metrics
	a.pollMetrics()
	a.reportMetrics(cResty)

	// And loop with timers
	for {
		select {
		case <-sigsEnd:
			close(a.workerPool)
			return
		case <-pollTicker.C:
			a.pollMetrics()
		case <-reportTicker.C:
			a.reportMetrics(cResty)
		}
	}
}

func gzipData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	_, err := gzipWriter.Write(data)
	if err != nil {
		return nil, err
	}
	err = gzipWriter.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
