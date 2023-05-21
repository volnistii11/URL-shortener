package server

import (
	"github.com/gin-gonic/gin"
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/handlers"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
)

type Server struct {
}

func RunServer(repository storage.Repository) {
	config.ParseFlags()

	h := handlers.NewHandlerProvider(repository)

	r := gin.Default()
	r.POST("/", h.CreateShortURL)
	r.GET("/:short_url", h.GetFullURL)

	r.Run(config.Addresses.Server)
}
