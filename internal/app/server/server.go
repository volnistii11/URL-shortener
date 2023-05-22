package server

import (
	"github.com/gin-gonic/gin"
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/handlers"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
)

type Router interface {
	RunServer(storage.Repository, config.Flags)
}

func NewRouter() Router {
	return &server{
		httpServer: gin.Default(),
	}
}

type server struct {
	httpServer *gin.Engine
}

func (srv *server) RunServer(repository storage.Repository, cfg config.Flags) {
	cfg.ParseFlags()

	h := handlers.NewHandlerProvider(repository, cfg)

	srv.httpServer.POST("/", h.CreateShortURL)
	srv.httpServer.GET("/:short_url", h.GetFullURL)

	srv.httpServer.Run(cfg.GetServer())
}
