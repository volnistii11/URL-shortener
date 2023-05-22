package handlers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	"net/http"
)

type HandlerProvider interface {
	CreateShortURL(ctx *gin.Context)
	GetFullURL(ctx *gin.Context)
}

func NewHandlerProvider(repository storage.Repository, cfg config.Flags) HandlerProvider {
	return &handlerURL{
		repo:  repository,
		flags: cfg,
	}
}

type handlerURL struct {
	repo  storage.Repository
	flags config.Flags
}

func (h *handlerURL) CreateShortURL(ctx *gin.Context) {
	ctx.Header("content-type", "text/plain; charset=utf-8")
	body, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if len(body) == 0 {
		err = errors.New("body is empty")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}

	shortURL, err := h.repo.WriteURL(string(body))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	respondingServerAddress := scheme + "://" + ctx.Request.Host + ctx.Request.RequestURI
	if h.flags.GetRespondingServer() != "" {
		respondingServerAddress = h.flags.GetRespondingServer() + "/"
	}

	fmt.Println(respondingServerAddress)
	ctx.String(http.StatusCreated, "%v%v", respondingServerAddress, shortURL)
}

func (h *handlerURL) GetFullURL(ctx *gin.Context) {
	shortURL := ctx.Params.ByName("short_url")

	fullURL, err := h.repo.ReadURL(shortURL)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	ctx.Header("Location", fullURL)
	ctx.Status(http.StatusTemporaryRedirect)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
