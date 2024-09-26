package agent

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"runtime"

	"github.com/go-resty/resty/v2"
	"github.com/melkomukovki/go-musthave-metrics/internal/config"
	"github.com/melkomukovki/go-musthave-metrics/internal/storage"
)

type Agent struct {
	pollCounter int64
	metrics     []storage.Metrics
	config      *config.ClientConfig
}

func NewAgent(cfg *config.ClientConfig) *Agent {
	return &Agent{
		config: cfg,
	}
}

func (a *Agent) addPollCounter() {
	a.pollCounter++
}

func (a *Agent) resetPollCounter() {
	a.pollCounter = 0
}

func (a *Agent) GetPollInterval() int {
	return a.config.PollInterval
}

func (a *Agent) GetReportInterval() int {
	return a.config.ReportInterval
}

func (a *Agent) PollMetrics() {
	a.addPollCounter()
	var newAr []storage.Metrics
	var rtm runtime.MemStats

	runtime.ReadMemStats(&rtm)

	allocValue := float64(rtm.Alloc)
	buckHashSys := float64(rtm.BuckHashSys)
	frees := float64(rtm.Frees)
	gCCPUFraction := float64(rtm.GCCPUFraction)
	gCSys := float64(rtm.GCSys)
	heapAlloc := float64(rtm.HeapAlloc)
	heapIdle := float64(rtm.HeapIdle)
	heapInuse := float64(rtm.HeapInuse)
	heapObjects := float64(rtm.HeapObjects)
	heapReleased := float64(rtm.HeapReleased)
	heapSys := float64(rtm.HeapSys)
	lastGC := float64(rtm.LastGC)
	lookups := float64(rtm.Lookups)
	mCacheInuse := float64(rtm.MCacheInuse)
	mCacheSys := float64(rtm.MCacheSys)
	mSpanInuse := float64(rtm.MSpanInuse)
	mSpanSys := float64(rtm.MSpanSys)
	mallocs := float64(rtm.Mallocs)
	nextGC := float64(rtm.NextGC)
	numForcedGC := float64(rtm.NumForcedGC)
	numGC := float64(rtm.NumGC)
	otherSys := float64(rtm.OtherSys)
	pauseTotalNs := float64(rtm.PauseTotalNs)
	stackInuse := float64(rtm.StackInuse)
	stackSys := float64(rtm.StackSys)
	sys := float64(rtm.Sys)
	totalAlloc := float64(rtm.TotalAlloc)
	randomValue := rand.Float64()

	newAr = append(newAr, storage.Metrics{ID: "Alloc", MType: storage.Gauge, Value: &allocValue})
	newAr = append(newAr, storage.Metrics{ID: "BuckHashSys", MType: storage.Gauge, Value: &buckHashSys})
	newAr = append(newAr, storage.Metrics{ID: "Frees", MType: storage.Gauge, Value: &frees})
	newAr = append(newAr, storage.Metrics{ID: "GCCPUFraction", MType: storage.Gauge, Value: &gCCPUFraction})
	newAr = append(newAr, storage.Metrics{ID: "GCSys", MType: storage.Gauge, Value: &gCSys})
	newAr = append(newAr, storage.Metrics{ID: "HeapAlloc", MType: storage.Gauge, Value: &heapAlloc})
	newAr = append(newAr, storage.Metrics{ID: "HeapIdle", MType: storage.Gauge, Value: &heapIdle})
	newAr = append(newAr, storage.Metrics{ID: "HeapInuse", MType: storage.Gauge, Value: &heapInuse})
	newAr = append(newAr, storage.Metrics{ID: "HeapObjects", MType: storage.Gauge, Value: &heapObjects})
	newAr = append(newAr, storage.Metrics{ID: "HeapReleased", MType: storage.Gauge, Value: &heapReleased})
	newAr = append(newAr, storage.Metrics{ID: "HeapSys", MType: storage.Gauge, Value: &heapSys})
	newAr = append(newAr, storage.Metrics{ID: "LastGC", MType: storage.Gauge, Value: &lastGC})
	newAr = append(newAr, storage.Metrics{ID: "Lookups", MType: storage.Gauge, Value: &lookups})
	newAr = append(newAr, storage.Metrics{ID: "MCacheInuse", MType: storage.Gauge, Value: &mCacheInuse})
	newAr = append(newAr, storage.Metrics{ID: "MCacheSys", MType: storage.Gauge, Value: &mCacheSys})
	newAr = append(newAr, storage.Metrics{ID: "MSpanInuse", MType: storage.Gauge, Value: &mSpanInuse})
	newAr = append(newAr, storage.Metrics{ID: "MSpanSys", MType: storage.Gauge, Value: &mSpanSys})
	newAr = append(newAr, storage.Metrics{ID: "Mallocs", MType: storage.Gauge, Value: &mallocs})
	newAr = append(newAr, storage.Metrics{ID: "NextGC", MType: storage.Gauge, Value: &nextGC})
	newAr = append(newAr, storage.Metrics{ID: "NumForcedGC", MType: storage.Gauge, Value: &numForcedGC})
	newAr = append(newAr, storage.Metrics{ID: "NumGC", MType: storage.Gauge, Value: &numGC})
	newAr = append(newAr, storage.Metrics{ID: "OtherSys", MType: storage.Gauge, Value: &otherSys})
	newAr = append(newAr, storage.Metrics{ID: "PauseTotalNs", MType: storage.Gauge, Value: &pauseTotalNs})
	newAr = append(newAr, storage.Metrics{ID: "StackInuse", MType: storage.Gauge, Value: &stackInuse})
	newAr = append(newAr, storage.Metrics{ID: "StackSys", MType: storage.Gauge, Value: &stackSys})
	newAr = append(newAr, storage.Metrics{ID: "Sys", MType: storage.Gauge, Value: &sys})
	newAr = append(newAr, storage.Metrics{ID: "TotalAlloc", MType: storage.Gauge, Value: &totalAlloc})
	newAr = append(newAr, storage.Metrics{ID: "RandomValue", MType: storage.Gauge, Value: &randomValue})
	newAr = append(newAr, storage.Metrics{ID: "PollCount", MType: storage.Counter, Delta: &a.pollCounter})

	a.metrics = newAr
}

func (a *Agent) ReportMetrics(c *resty.Client) {
	url := fmt.Sprintf("http://%s/update/", a.config.Address)
	for _, metric := range a.metrics {
		mMarshaled, _ := json.Marshal(metric)
		resp, err := c.R().
			SetBody(mMarshaled).
			SetHeader("Content-Type", "application/json").
			Post(url)
		if err != nil {
			fmt.Printf("Detected error: %s\n", err.Error())
		}
		fmt.Printf("Metric %s sended. Status code: %d\n", metric.ID, resp.StatusCode())
	}
	a.resetPollCounter()
}
