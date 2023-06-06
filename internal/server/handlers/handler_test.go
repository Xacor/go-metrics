package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Xacor/go-metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
)

func TestAPI_UpdateHandler(t *testing.T) {
	type want struct {
		statusCode int
	}

	tests := []struct {
		name string
		repo storage.MetricRepo
		url  string
		want want
	}{
		{
			name: "update counter",
			repo: storage.NewMemStorage(),
			url:  "/update/counter/someName/123",
			want: want{http.StatusOK},
		},
		{
			name: "counter invalid value",
			repo: storage.NewMemStorage(),
			url:  "/update/counter/someName/123.123",
			want: want{http.StatusBadRequest},
		},
		{
			name: "update gauge",
			repo: storage.NewMemStorage(),
			url:  "/update/gauge/someName/123.123",
			want: want{http.StatusOK},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &API{
				Repo: tt.repo,
			}
			request := httptest.NewRequest(http.MethodPost, tt.url, nil)
			w := httptest.NewRecorder()
			api.UpdateHandler(w, request)

			result := w.Result()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			result.Body.Close()
		})
	}
}
