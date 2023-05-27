package api

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	"net/http"
	"net/http/httptest"
	"testing"
)

func SetUpRouter() *gin.Engine {
	router := gin.Default()
	return router
}

func TestCreateShortURL(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}
	type request struct {
		Url string `json:"url,omitempty"`
	}
	tests := []struct {
		name    string
		request string
		url     request
		want    want
	}{
		{
			name:    "positive test #1",
			request: "/api/shorten",
			url: request{
				Url: "https://practicum.yandex.ru",
			},
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name:    "negative test - body is empty",
			request: "/api/shorten",
			url: request{
				Url: "",
			},
			want: want{
				code:        400,
				contentType: "application/json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(tt.url)

			repo := storage.NewRepository()
			flags := config.NewFlags()
			api := NewAPIServiceServer(repo, flags)

			r := SetUpRouter()
			r.POST("/api/shorten", api.CreateShortURL)
			req, _ := http.NewRequest(http.MethodPost, tt.request, &buf)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.want.code, w.Code)
			assert.Equal(t, tt.want.contentType, w.Header().Get("Content-Type"))
		})
	}
}
