package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func SetUpRouter() *gin.Engine {
	router := gin.Default()
	return router
}

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
				contentType:            "text/plain; charset=utf-8",
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

			r := SetUpRouter()
			r.POST("/", CreateShortURL)
			req, _ := http.NewRequest(http.MethodPost, tt.request, bodyReader)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.want.code, w.Code)
			assert.Equal(t, tt.want.contentType, w.Header().Get("Content-Type"))

			resBody, err := io.ReadAll(w.Body)
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
			r := SetUpRouter()
			r.POST("/:short_url", GetFullURL)
			req, _ := http.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.want.code, w.Code)
			assert.Equal(t, tt.want.locationIsNotEmpty, len(w.Header().Get("Location")) > 0)
		})
	}
}
