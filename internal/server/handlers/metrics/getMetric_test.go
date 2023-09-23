package metrics

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Xacor/go-metrics/internal/logger"
	mock_storage "github.com/Xacor/go-metrics/internal/server/mocks"
	"github.com/Xacor/go-metrics/internal/server/model"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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

			assert.Equal(t, tt.want.code, w.Result().StatusCode)
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
	m.EXPECT().All(gomock.Any()).Return([]model.Metrics{{"counter1", model.TypeCounter, &val, nil}}, nil).AnyTimes()

	logger.Initialize("DEBUG")
	l := logger.Get()

	api := &API{
		repo:   m,
		logger: l,
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
				assert.Equal(b, bm.want.code, w.Result().StatusCode)
				b.StartTimer()
			}
		})

	}
}
