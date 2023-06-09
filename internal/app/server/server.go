package server

import (
	"github.com/volnistii11/URL-shortener/internal/app/api"
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/handlers"
	"github.com/volnistii11/URL-shortener/internal/app/middlewares"
	"github.com/volnistii11/URL-shortener/internal/app/storage"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Runner interface {
	Router(storage.Repository, config.Flags) *gin.Engine
}

func NewRouter(logger *zap.Logger) Runner {
	return &server{
		httpServer: gin.New(),
		logger:     logger,
	}
}

type server struct {
	httpServer *gin.Engine
	logger     *zap.Logger
}

func (srv *server) Router(repository storage.Repository, cfg config.Flags) *gin.Engine {
	h := handlers.NewHandlerProvider(repository, cfg)
	m := middlewares.NewMiddlewareProvider(srv.logger)
	a := api.NewAPIServiceServer(repository, cfg)

	srv.httpServer.Use(gin.Recovery())
	srv.httpServer.Use(m.LogHTTPHandler())
	srv.httpServer.Use(m.GZIPHandler())
	srv.httpServer.Use(m.Auth())
	srv.httpServer.POST("/", h.CreateShortURL)
	srv.httpServer.GET("/:short_url", h.GetFullURL)
	srv.httpServer.GET("/ping", h.PingDatabaseServer)

	srv.httpServer.POST("/api/shorten", a.CreateShortURL)
	srv.httpServer.POST("/api/shorten/batch", a.CreateShortURLBatch)
	srv.httpServer.GET("/api/user/urls", a.GetAllUserURLS)
	srv.httpServer.DELETE("api/user/urls", a.DeleteUserURLS)
	return srv.httpServer
}
