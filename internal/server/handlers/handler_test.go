package handlers

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Xacor/go-metrics/internal/logger"
	"github.com/Xacor/go-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestAPI_UpdateRouter(t *testing.T) {
	err := logger.Initialize("debug")
	if err != nil {
		log.Println(err)
	}

	api := NewAPI(storage.NewMemStorage(), logger.Get())

	r := chi.NewRouter()
	r.Post("/update/{metricType}/{metricID}/{metricValue}", api.UpdateHandler)
	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		statusCode int
	}

	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "update counter",
			url:  "/update/counter/someName/123",
			want: want{http.StatusOK},
		},
		{
			name: "counter invalid value",
			url:  "/update/counter/someName/123.123",
			want: want{http.StatusBadRequest},
		},

		{
			name: "counter without id",
			url:  "/update/counter/123",
			want: want{http.StatusNotFound},
		},
		{
			name: "update gauge",
			url:  "/update/gauge/someName1/123.123",
			want: want{http.StatusOK},
		},
		{
			name: "gauge invalid value",
			url:  "/update/gauge/someName1/zxc",
			want: want{http.StatusBadRequest},
		},
		{
			name: "gauge without id",
			url:  "/update/gauge/123",
			want: want{http.StatusNotFound},
		},
		{
			name: "invalid type",
			url:  "/update/unknown/anyName/123",
			want: want{http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := testRequest(t, ts, "POST", tt.url)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			resp.Body.Close()
		})
	}
}
