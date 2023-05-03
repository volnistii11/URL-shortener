package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCreateShortURL(t *testing.T) {
	type want struct {
		code                   int
		responseBodyIsNotEmpty bool
		contentType            string
	}
	tests := []struct {
		name        string
		request     string
		requestBody string
		want        want
	}{
		{
			name:        "positive test #1",
			request:     "/",
			requestBody: "https://practicum.yandex.ru/",
			want: want{
				code:                   201,
				responseBodyIsNotEmpty: true,
				contentType:            "text/plain",
			},
		},
		{
			name:        "negative test - body is empty",
			request:     "/",
			requestBody: "",
			want: want{
				code:                   400,
				responseBodyIsNotEmpty: false,
				contentType:            "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyReader := strings.NewReader(tt.requestBody)
			request := httptest.NewRequest(http.MethodPost, tt.request, bodyReader)

			w := httptest.NewRecorder()
			CreateShortURL(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			bodyIsNotEmpty := true
			_, err = url.ParseRequestURI(string(resBody))
			if err != nil {
				bodyIsNotEmpty = false
			}
			assert.Equal(t, tt.want.responseBodyIsNotEmpty, bodyIsNotEmpty)
		})
	}
}

func TestGetFullURL(t *testing.T) {
	type want struct {
		code               int
		locationIsNotEmpty bool
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "positive test #1",
			request: "http://localhost:8080/sKtBWabUkV",
			want: want{
				code:               307,
				locationIsNotEmpty: true,
			},
		},
		{
			name:    "negative test - no short URL",
			request: "http://localhost:8080/fail",
			want: want{
				code:               400,
				locationIsNotEmpty: false,
			},
		},
	}

	//Это надо будет переписать, когда разберусь и перепишу storage
	storage.URLMap = map[string]string{}
	storage.URLMap["sKtBWabUkV"] = "https://go.dev/tour/welcome/1"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()
			GetFullURL(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.code, result.StatusCode)
			assert.Equal(t, tt.want.locationIsNotEmpty, len(result.Header.Get("Location")) > 0)
		})
	}
}
