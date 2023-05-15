package server

import (
	"github.com/gin-gonic/gin"
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/handlers"
)

func RunServer() {
	config.ParseFlags()

	r := gin.Default()
	r.POST("/", handlers.CreateShortURL)
	r.GET("/:short_url", handlers.GetFullURL)

	r.Run(config.Addresses.Server)
}
