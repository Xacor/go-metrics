package database

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mock_storage "github.com/Xacor/go-metrics/internal/server/mocks"
	"github.com/go-playground/assert/v2"
	"github.com/golang/mock/gomock"
)

func TestHealthService_Ping(t *testing.T) {
	type fields struct {
		storage *mock_storage.MockPinger
	}
	type want struct {
		code int
	}
	tests := []struct {
		name    string
		want    want
		prepare func(f *fields)
	}{
		{
			name: "ping_ok",
			want: want{
				code: http.StatusOK,
			},
			prepare: func(f *fields) {
				f.storage.EXPECT().Ping(gomock.Any())
			},
		},
		{
			name: "db_error",
			want: want{
				code: http.StatusInternalServerError,
			},
			prepare: func(f *fields) {
				f.storage.EXPECT().Ping(gomock.Any()).Return(errors.New("db_error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := fields{
				storage: mock_storage.NewMockPinger(ctrl),
			}

			api := &HealthService{
				db: f.storage,
			}

			r := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()

			if tt.prepare != nil {
				tt.prepare(&f)
			}

			api.Ping(w, r)

			assert.Equal(t, tt.want.code, w.Result().StatusCode)
			w.Result().Body.Close()

		})
	}
}
