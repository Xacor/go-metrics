package metrics

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Xacor/go-metrics/internal/logger"
	mock_storage "github.com/Xacor/go-metrics/internal/server/mocks"
	"github.com/Xacor/go-metrics/internal/server/model"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAPI_MetricsHandler(t *testing.T) {
	type fields struct {
		storage *mock_storage.MockStorage
	}
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name    string
		want    want
		prepare func(f *fields)
	}{
		{
			name: "positive_case",
			want: want{
				code:        http.StatusOK,
				contentType: "text/html",
			},
			prepare: func(f *fields) {
				var val int64 = 1
				f.storage.EXPECT().All(gomock.Any()).Return([]model.Metrics{{
					Name:  "counter1",
					MType: model.TypeCounter,
					Delta: &val,
					Value: nil,
				}},
					nil)
			},
		},
		{
			name: "db_error_case",
			want: want{
				code:        http.StatusInternalServerError,
				contentType: "text/html",
			},
			prepare: func(f *fields) {
				f.storage.EXPECT().All(gomock.Any()).Return(nil, errors.New("some error")).AnyTimes()
			},
		},
	}

	logger.Initialize("DEBUG")
	l := logger.Get()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			f := fields{
				storage: mock_storage.NewMockStorage(ctrl),
			}

			if tt.prepare != nil {
				tt.prepare(&f)
			}

			api := &API{
				repo:   f.storage,
				logger: l,
			}

			api.MetricsHandler(w, r)
			resp := w.Result()
			defer resp.Body.Close()
			assert.Equal(t, tt.want.code, resp.StatusCode)
		})
	}
}

func BenchmarkAPI_MetricsHandler(b *testing.B) {
	type want struct {
		code        int
		contentType string
	}
	benchmarks := []struct {
		name string
		want want
	}{
		{
			name: "bench metrics handler",
			want: want{
				code:        http.StatusOK,
				contentType: "text/html",
			},
		},
	}

	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	m := mock_storage.NewMockStorage(ctrl)
	var val int64 = 1
	m.EXPECT().All(gomock.Any()).Return([]model.Metrics{{Name: "counter1", MType: model.TypeCounter, Delta: &val, Value: nil}}, nil).AnyTimes()

	api := &API{
		repo:   m,
		logger: zap.NewNop(),
	}

	b.ResetTimer()
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				w := httptest.NewRecorder()
				b.StartTimer()

				api.MetricsHandler(w, r)

				b.StopTimer()
				resp := w.Result()
				resp.Body.Close()
				assert.Equal(b, bm.want.code, resp.StatusCode)
				b.StartTimer()
			}
		})

	}
}

func BenchmarkAPI_MetricJSON(b *testing.B) {
	type fields struct {
		storage *mock_storage.MockStorage
	}
	type want struct {
		code int
	}
	benchmarks := []struct {
		name    string
		body    []byte
		want    want
		prepare func(f *fields)
	}{
		{
			name: "benchmark",
			body: []byte(`{"id": "name1","type": "counter","delta": 1}`),
			want: want{
				code: http.StatusOK,
			},
			prepare: func(f *fields) {
				var val int64 = 1
				f.storage.EXPECT().Get(gomock.Any(), "name1").Return(model.Metrics{
					Name:  "name1",
					MType: model.TypeCounter,
					Delta: &val,
					Value: nil,
				},
					nil)
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				ctrl := gomock.NewController(b)
				defer ctrl.Finish()

				f := fields{
					storage: mock_storage.NewMockStorage(ctrl),
				}

				api := &API{
					repo:   f.storage,
					logger: zap.NewNop(),
				}

				r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bm.body))
				r.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				if bm.prepare != nil {
					bm.prepare(&f)
				}
				b.StartTimer()

				api.MetricJSON(w, r)

				b.StopTimer()
				resp := w.Result()
				resp.Body.Close()
				b.StartTimer()

			}
		})
	}
}

func TestAPI_MetricJSON(t *testing.T) {
	type fields struct {
		storage *mock_storage.MockStorage
	}
	type want struct {
		code int
	}
	tests := []struct {
		name    string
		body    []byte
		want    want
		prepare func(f *fields)
	}{
		{
			name: "positive",
			body: []byte(`{"id": "name1","type": "counter","delta": 1}`),
			want: want{
				code: http.StatusOK,
			},
			prepare: func(f *fields) {
				var val int64 = 1
				f.storage.EXPECT().Get(gomock.Any(), "name1").Return(model.Metrics{
					Name:  "name1",
					MType: model.TypeCounter,
					Delta: &val,
					Value: nil,
				},
					nil)
			},
		},
		{
			name: "db_error",
			body: []byte(`{"id": "name1","type": "counter","delta": 1}`),
			want: want{
				code: http.StatusNotFound,
			},
			prepare: func(f *fields) {
				f.storage.EXPECT().Get(gomock.Any(), "name1").Return(model.Metrics{}, errors.New("db error"))
			},
		},
		{
			name: "invalid_body",
			body: []byte(`{"id": "name1","type": "counter","delta": 1`),
			want: want{
				code: http.StatusBadRequest,
			},
			prepare: nil,
		},
	}

	l := logger.Get()
	for _, bm := range tests {
		t.Run(bm.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := fields{
				storage: mock_storage.NewMockStorage(ctrl),
			}

			api := &API{
				repo:   f.storage,
				logger: l,
			}

			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bm.body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			if bm.prepare != nil {
				bm.prepare(&f)
			}

			api.MetricJSON(w, r)
			resp := w.Result()
			defer resp.Body.Close()
			assert.Equal(t, bm.want.code, resp.StatusCode)
		})
	}
}
