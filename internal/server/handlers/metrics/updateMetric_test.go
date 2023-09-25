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

func TestAPI_UpdateJSON(t *testing.T) {
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
			name: "new_metric",
			body: []byte(`{"id": "name1","type": "counter","delta": 1}`),
			want: want{
				code: http.StatusOK,
			},
			prepare: func(f *fields) {
				f.storage.EXPECT().Get(gomock.Any(), "name1").Return(model.Metrics{}, errors.New("no rows"))
				var val int64 = 1
				metric := model.Metrics{
					Name:  "name1",
					MType: model.TypeCounter,
					Delta: &val,
					Value: nil,
				}
				f.storage.EXPECT().Create(gomock.Any(), metric).Return(metric, nil)
			},
		},
		{
			name: "old_metric",
			body: []byte(`{"id": "name1","type": "counter","delta": 1}`),
			want: want{
				code: http.StatusOK,
			},
			prepare: func(f *fields) {

				var val int64 = 1
				metric := model.Metrics{
					Name:  "name1",
					MType: model.TypeCounter,
					Delta: &val,
					Value: nil,
				}
				f.storage.EXPECT().Get(gomock.Any(), "name1").Return(metric, nil)
				f.storage.EXPECT().Update(gomock.Any(), metric).Return(metric, nil)
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

			r := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(bm.body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			if bm.prepare != nil {
				bm.prepare(&f)
			}

			api.UpdateJSON(w, r)

			assert.Equal(t, bm.want.code, w.Result().StatusCode)
			w.Result().Body.Close()

		})
	}
}

func TestAPI_UpdateMetrics(t *testing.T) {
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
			name: "new_metric",
			body: []byte(`[{"id": "name1","type": "counter","delta": 1}]`),
			want: want{
				code: http.StatusOK,
			},
			prepare: func(f *fields) {
				var val int64 = 1
				metrics := []model.Metrics{{
					Name:  "name1",
					MType: model.TypeCounter,
					Delta: &val,
					Value: nil,
				}}
				f.storage.EXPECT().UpdateBatch(gomock.Any(), metrics).Return(nil)
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
		{
			name: "db_error",
			body: []byte(`[{"id": "name1","type": "counter","delta": 1}]`),
			want: want{
				code: http.StatusInternalServerError,
			},
			prepare: func(f *fields) {
				var val int64 = 1
				metrics := []model.Metrics{{
					Name:  "name1",
					MType: model.TypeCounter,
					Delta: &val,
					Value: nil,
				}}
				f.storage.EXPECT().UpdateBatch(gomock.Any(), metrics).Return(errors.New("db error"))
			},
		},
	}

	l := logger.Get()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := fields{
				storage: mock_storage.NewMockStorage(ctrl),
			}

			api := &API{
				repo:   f.storage,
				logger: l,
			}

			r := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			if tt.prepare != nil {
				tt.prepare(&f)
			}

			api.UpdateMetrics(w, r)

			assert.Equal(t, tt.want.code, w.Result().StatusCode)
			w.Result().Body.Close()

		})
	}
}

func BenchmarkAPI_UpdateMetrics(b *testing.B) {
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
			name: "new_metric",
			body: []byte(`[{"id": "name1","type": "counter","delta": 1}, {"id": "name2","type": "gauge","value": 2.0}]`),
			want: want{
				code: http.StatusOK,
			},
			prepare: func(f *fields) {
				f.storage.EXPECT().UpdateBatch(gomock.Any(), gomock.Any()).Return(nil)
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
		{
			name: "db_error",
			body: []byte(`[{"id": "name1","type": "counter","delta": 1}]`),
			want: want{
				code: http.StatusInternalServerError,
			},
			prepare: func(f *fields) {
				var val int64 = 1
				metrics := []model.Metrics{{
					Name:  "name1",
					MType: model.TypeCounter,
					Delta: &val,
					Value: nil,
				}}
				f.storage.EXPECT().UpdateBatch(gomock.Any(), metrics).Return(errors.New("db error"))
			},
		},
	}

	b.ResetTimer()
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

				r := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(bm.body))
				r.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				if bm.prepare != nil {
					bm.prepare(&f)
				}
				b.StartTimer()
				api.UpdateMetrics(w, r)

				w.Result().Body.Close()
			}
		})
	}
}
