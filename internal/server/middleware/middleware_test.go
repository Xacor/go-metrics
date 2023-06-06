package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPost(t *testing.T) {
	type want struct {
		code int
	}

	tests := []struct {
		name   string
		method string
		want   want
	}{
		{
			name:   "positive test",
			method: http.MethodPost,
			want:   want{http.StatusOK},
		},
		{
			name:   "negative test",
			method: http.MethodGet,
			want:   want{http.StatusMethodNotAllowed},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/update/counter/someName/123", nil)
			w := httptest.NewRecorder()
			next := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}
			Post(http.HandlerFunc(next)).ServeHTTP(w, request)

			res := w.Result()

			assert.Equal(t, res.StatusCode, test.want.code)
			res.Body.Close()
		})
	}

}

func TestTextPlain(t *testing.T) {
	type want struct {
		code int
	}

	tests := []struct {
		name        string
		contentType string
		want        want
	}{
		{
			name:        "positive test",
			contentType: "text/plain",
			want:        want{http.StatusOK},
		},
		{
			name:        "negative test",
			contentType: "application/json",
			want:        want{http.StatusBadRequest},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/update/counter/someName/123", nil)
			request.Header.Add("Content-Type", test.contentType)
			w := httptest.NewRecorder()
			next := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}
			TextPlain(http.HandlerFunc(next)).ServeHTTP(w, request)

			res := w.Result()

			assert.Equal(t, res.StatusCode, test.want.code)
			res.Body.Close()
		})
	}
}

func TestValidateParams(t *testing.T) {
	type want struct {
		code int
	}

	tests := []struct {
		name string
		path string
		want want
	}{
		{
			name: "update",
			path: "/update/counter/someName/123",
			want: want{http.StatusOK},
		},
		{
			name: "invalid type",
			path: "/update/unknown/someName/123",
			want: want{http.StatusBadRequest},
		},
		{
			name: "invalid value",
			path: "/update/counter/someName/zxczxc",
			want: want{http.StatusBadRequest},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.path, nil)
			w := httptest.NewRecorder()
			next := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}
			ValidateParams(http.HandlerFunc(next)).ServeHTTP(w, request)

			res := w.Result()

			assert.Equal(t, res.StatusCode, test.want.code)
			res.Body.Close()
		})
	}
}
