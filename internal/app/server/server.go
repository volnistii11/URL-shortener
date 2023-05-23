package server

import (
	"github.com/gin-gonic/gin"
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/handlers"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
)

type Runner interface {
	Router(storage.Repository, config.Flags) *gin.Engine
}

func NewRouter() Runner {
	return &server{
		httpServer: gin.Default(),
	}
}

type server struct {
	httpServer *gin.Engine
}

func (srv *server) Router(repository storage.Repository, cfg config.Flags) *gin.Engine {
	h := handlers.NewHandlerProvider(repository, cfg)
	srv.httpServer.POST("/", h.CreateShortURL)
	srv.httpServer.GET("/:short_url", h.GetFullURL)
	return srv.httpServer
}
