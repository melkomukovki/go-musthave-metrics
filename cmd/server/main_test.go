package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func SetUpRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/update/:mType/:mName/:mValue", PostMetricHandler)
	r.GET("/value/:mType/:mName", GetMetricHandler)
	r.GET("/", ShowMetrics)
	return r
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

	r := SetUpRouter()

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
