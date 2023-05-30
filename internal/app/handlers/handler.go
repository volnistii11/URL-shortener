package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	file_storage "github.com/volnistii11/URL-shortener/internal/app/storage/file"
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

	var originalURL string
	if h.flags.GetFileStoragePath() == "" {
		originalURL = string(body)
	} else {
		Producer, _ := file_storage.NewProducer(h.flags.GetFileStoragePath())
		defer Producer.Close()
		bufEvent := file_storage.Event{}
		_ = json.Unmarshal(body, &bufEvent)
		Producer.WriteEvent(&bufEvent)
		originalURL = bufEvent.OriginalURL
		if len(originalURL) == 0 {
			originalURL = string(body)
		}
	}
	shortURL, err2 := h.repo.WriteURL(originalURL)
	if err2 != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err2))
		return
	}

	respondingServerAddress := scheme + "://" + ctx.Request.Host + ctx.Request.RequestURI
	if h.flags.GetRespondingServer() != "" {
		respondingServerAddress = h.flags.GetRespondingServer() + "/"
	}

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
