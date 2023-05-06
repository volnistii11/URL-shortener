package server

import (
	"github.com/gin-gonic/gin"
	"github.com/volnistii11/URL-shortener/internal/app/handlers"
)

func RunServer() {

	r := gin.Default()
	r.POST("/", handlers.CreateShortURL)
	r.GET("/:short_url", handlers.GetFullURL)

	err := r.Run("localhost:8080")
	if err != nil {
		panic(err)
	}
}
