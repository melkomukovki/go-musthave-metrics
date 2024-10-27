package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/melkomukovki/go-musthave-metrics/internal/server"
	"github.com/melkomukovki/go-musthave-metrics/internal/server/config"
	"github.com/melkomukovki/go-musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestNewServer(t *testing.T) {
	server, err := server.New(&config.ServerConfig{
		Address:         ":8080",
		StoreInterval:   10,
		FileStoragePath: "/tmp/test.txt",
		Restore:         false,
		DataSourceName:  "",
		HashKey:         "",
	})
	require.NoError(t, err)
	require.NotNil(t, server)
}

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
			name: "Update counter metric",
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

	r, _ := server.New(&config.ServerConfig{
		Address:         ":8080",
		StoreInterval:   10,
		FileStoragePath: "/tmp/test.txt",
		Restore:         false,
		DataSourceName:  "",
		HashKey:         "",
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", tt.url, nil)
			req.Header.Add("Content-Type", "text/plain")
			r.Engine.ServeHTTP(w, req)
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

	r, _ := server.New(&config.ServerConfig{
		Address:         ":8080",
		StoreInterval:   10,
		FileStoragePath: "/tmp/test.txt",
		Restore:         false,
		DataSourceName:  "",
		HashKey:         "",
	})

	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "Get counter metric value",
			url:  "/value/counter/testCounter",
			want: want{
				code:        200,
				contentType: "text/plain",
				content:     "123",
			},
		},
		{
			name: "Get gauge metric value",
			url:  "/value/gauge/testGauge",
			want: want{
				code:        200,
				contentType: "text/plain",
				content:     "333.123",
			},
		},
		{
			name: "Get undefined metric",
			url:  "/value/gauge/noMetric",
			want: want{
				code:        404,
				contentType: "text/plain",
				content:     "metric not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tW := httptest.NewRecorder()
			for _, rd := range data {
				req, _ := http.NewRequest("POST", rd, nil)
				req.Header.Add("Content-Type", "text/plain")
				r.Engine.ServeHTTP(tW, req)
			}
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.url, nil)
			req.Header.Add("Content-Type", "text/plain")
			r.Engine.ServeHTTP(w, req)

			assert.Equal(t, tt.want.code, w.Code)
			assert.True(t, strings.Contains(w.Header().Get("Content-Type"), tt.want.contentType))
			assert.Equal(t, tt.want.content, w.Body.String())
		})
	}
}

func TestPingHadnler(t *testing.T) {
	r, _ := server.New(&config.ServerConfig{
		Address:         ":8080",
		StoreInterval:   10,
		FileStoragePath: "/tmp/test.txt",
		Restore:         false,
		DataSourceName:  "",
		HashKey:         "",
	})

	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/ping", nil)
	require.NoError(t, err)
	r.Engine.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "Pong", rec.Body.String())
}

func TestPostMetricHandlerJSON(t *testing.T) {
	type want struct {
		code        int
		contentType string
		hashHeader  string
	}

	var testGaugeValue = 123.123
	var testCounterValue int64 = 123
	var updateURL = "/update/"
	var hashKey = "testkey"

	tests := []struct {
		name   string
		metric storage.Metrics
		hash   string
		want   want
	}{
		{
			name: "Post gauge metric",
			metric: storage.Metrics{
				ID:    "test_gauge",
				MType: storage.Gauge,
				Value: &testGaugeValue,
			},
			hash: "766c1fc296f3a1f52fa1f06088c3807e1beee03bc8bec25731a6c0259a13c8d7",
			want: want{
				code:        200,
				contentType: "application/json",
				hashHeader:  "",
			},
		},
		{
			name: "Post counter metric",
			metric: storage.Metrics{
				ID:    "test_counter",
				MType: storage.Counter,
				Delta: &testCounterValue,
			},
			hash: "ea92a050e78be97ac2474399c6e30bb63c5706e750877632ad20063c050eee29",
			want: want{
				code:        200,
				contentType: "application/json",
			},
		},
		{
			name: "Post wrong type metric",
			metric: storage.Metrics{
				ID:    "test_wrong_type",
				MType: "unsupported_type",
			},
			hash: "4d32249b0caa562e7d0a832d5b8fc14f1d046e7083be0b33ce26e251a55c9d82",
			want: want{
				code:        400,
				contentType: "application/json",
			},
		},
		{
			name: "Post metric with wrong value",
			metric: storage.Metrics{
				ID:    "test_counter_with_value",
				MType: storage.Counter,
				Value: &testGaugeValue,
			},
			hash: "3e3917acfc003d9a8b68eac2aa32a09299d1a1d820b5a4cfbecb04dd8e816272",
			want: want{
				code:        400,
				contentType: "application/json",
			},
		},
		{
			name: "Post metric with wrong hash",
			metric: storage.Metrics{
				ID:    "test_gauge",
				MType: storage.Gauge,
				Value: &testGaugeValue,
			},
			hash: "d3JvbmdoYXNo",
			want: want{
				code:        400,
				contentType: "application/json",
			},
		},
	}

	r, _ := server.New(&config.ServerConfig{
		Address:         ":8080",
		StoreInterval:   10,
		FileStoragePath: "/tmp/test.txt",
		Restore:         false,
		DataSourceName:  "",
		HashKey:         hashKey,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			body, err := json.Marshal(tt.metric)
			require.NoError(t, err)

			compressedData, err := gzipData(body)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", updateURL, bytes.NewReader(compressedData))
			require.NoError(t, err)

			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("HashSHA256", tt.hash)
			req.Header.Add("Content-Encoding", "gzip")
			req.Header.Add("Accept-Encoding", "gzip")

			r.Engine.ServeHTTP(w, req)
			assert.Equal(t, tt.want.code, w.Code)
			assert.True(t, strings.Contains(w.Header().Get("Content-Type"), tt.want.contentType))
		})
	}
}
