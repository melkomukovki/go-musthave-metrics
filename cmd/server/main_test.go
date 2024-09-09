package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/melkomukovki/go-musthave-metrics/internal/server"
	"github.com/melkomukovki/go-musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestPostMetricHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}

	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "Positive test",
			url:  "/update/counter/testCounter/123",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
		},
		{
			name: "Request wrong url",
			url:  "/update/wrongURL",
			want: want{
				code:        404,
				contentType: "text/plain",
			},
		},
		{
			name: "Request invalid metric type",
			url:  "/update/metric/testMetric/123",
			want: want{
				code:        400,
				contentType: "text/plain",
			},
		},
	}

	var store storage.Storage = storage.MemStorage{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}

	r := server.NewServerRouter(store)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", tt.url, nil)
			req.Header.Add("Content-Type", "text/plain")
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.want.code, w.Code)
			assert.True(t, strings.Contains(w.Header().Get("Content-Type"), tt.want.contentType))
		})
	}

}

func TestGetMetricHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
		content     string
	}

	data := []string{
		"/update/counter/testCounter/123",
		"/update/gauge/testGauge/333.12345",
	}

	var store storage.Storage = storage.MemStorage{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}

	r := server.NewServerRouter(store)

	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "Positive Counter GET",
			url:  "/value/counter/testCounter",
			want: want{
				code:        200,
				contentType: "text/plain",
				content:     "123",
			},
		},
		{
			name: "Positive Gauge GET",
			url:  "/value/gauge/testGauge",
			want: want{
				code:        200,
				contentType: "text/plain",
				content:     "333.123",
			},
		},
		{
			name: "Negative GET",
			url:  "/value/gauge/noMetric",
			want: want{
				code:        404,
				contentType: "text/plain",
				content:     "Can't found metric",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tW := httptest.NewRecorder()
			for _, rd := range data {
				req, _ := http.NewRequest("POST", rd, nil)
				req.Header.Add("Content-Type", "text/plain")
				r.ServeHTTP(tW, req)
			}
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.url, nil)
			req.Header.Add("Content-Type", "text/plain")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.want.code, w.Code)
			assert.True(t, strings.Contains(w.Header().Get("Content-Type"), tt.want.contentType))

			bB, err := io.ReadAll(w.Result().Body)
			if err != nil {
				panic(err)
			}
			defer w.Result().Body.Close()
			assert.Equal(t, tt.want.content, string(bB))
		})
	}

}
