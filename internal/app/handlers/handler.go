package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	"github.com/volnistii11/URL-shortener/internal/app/storage/file"

	"github.com/gin-gonic/gin"
)

type HandlerProvider interface {
	CreateShortURL(ctx *gin.Context)
	GetFullURL(ctx *gin.Context)
	PingDatabaseServer(ctx *gin.Context)
	GetStorageType() string
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

	var shortURL string

	switch h.GetStorageType() {
	case "database":

	case "file":
		Producer, err := file.NewProducer(h.flags.GetFileStoragePath())
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		defer Producer.Close()
		bufEvent := file.Event{}
		err = json.Unmarshal(body, &bufEvent)
		if err != nil {
			bufEvent.OriginalURL = string(body)
			shortURL, err = h.repo.WriteURL(bufEvent.OriginalURL)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, errorResponse(err))
				return
			}
			bufEvent.ShortURL = shortURL
		} else {
			shortURL = bufEvent.ShortURL
		}
		Producer.WriteEvent(&bufEvent)
	case "memory":
		originalURL := string(body)
		shortURL, err = h.repo.WriteURL(originalURL)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
	}
	//if h.flags.GetFileStoragePath() == "" {
	//	originalURL := string(body)
	//	shortURL, err = h.repo.WriteURL(originalURL)
	//	if err != nil {
	//		ctx.JSON(http.StatusBadRequest, errorResponse(err))
	//		return
	//	}
	//} else {
	//	Producer, err := file.NewProducer(h.flags.GetFileStoragePath())
	//	if err != nil {
	//		ctx.JSON(http.StatusBadRequest, errorResponse(err))
	//		return
	//	}
	//	defer Producer.Close()
	//	bufEvent := file.Event{}
	//	err = json.Unmarshal(body, &bufEvent)
	//	if err != nil {
	//		bufEvent.OriginalURL = string(body)
	//		shortURL, err = h.repo.WriteURL(bufEvent.OriginalURL)
	//		if err != nil {
	//			ctx.JSON(http.StatusBadRequest, errorResponse(err))
	//			return
	//		}
	//		bufEvent.ShortURL = shortURL
	//	} else {
	//		shortURL = bufEvent.ShortURL
	//	}
	//	Producer.WriteEvent(&bufEvent)
	//}

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

func (h *handlerURL) PingDatabaseServer(ctx *gin.Context) {
	if err := h.repo.GetDatabase().Ping(); err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	ctx.Status(http.StatusOK)
}

func (h *handlerURL) GetStorageType() string {
	if h.flags.GetDatabaseDSN() != "" {
		return "database"
	} else if h.flags.GetFileStoragePath() != "" {
		return "file"
	} else {
		return "memory"
	}
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
