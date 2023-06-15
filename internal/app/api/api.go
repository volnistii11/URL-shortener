package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/volnistii11/URL-shortener/internal/app/storage/database"
	"github.com/volnistii11/URL-shortener/internal/app/storage/file"
	"github.com/volnistii11/URL-shortener/internal/app/utils"
	"math/rand"
	"net/http"

	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/storage"

	"github.com/gin-gonic/gin"
)

type Provider interface {
	CreateShortURL(ctx *gin.Context)
	CreateShortURLBatch(ctx *gin.Context)
}

func NewAPIServiceServer(repository storage.Repository, cfg config.Flags) Provider {
	return &api{
		repo:  repository,
		flags: cfg,
	}
}

type api struct {
	repo  storage.Repository
	flags config.Flags
}

type request struct {
	URL string `json:"url,omitempty"`
}

type response struct {
	Result string `json:"result,omitempty"`
}

func (a *api) CreateShortURL(ctx *gin.Context) {
	ctx.Header("content-type", "application/json")
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
	bufRequest := request{}
	if err = json.Unmarshal(body, &bufRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	shortURL, err := a.repo.WriteURL(bufRequest.URL)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	respondingServerAddress := fmt.Sprintf("%v://%v/", scheme, ctx.Request.Host)
	if a.flags.GetRespondingServer() != "" {
		respondingServerAddress = fmt.Sprintf("%v/", a.flags.GetRespondingServer())
	}
	buffResponse := response{
		Result: fmt.Sprintf("%v%v", respondingServerAddress, shortURL),
	}
	ctx.JSON(http.StatusCreated, buffResponse)
}

func (a *api) CreateShortURLBatch(ctx *gin.Context) {
	ctx.Header("content-type", "application/json")
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

	var urls []storage.URLStorage
	if err = json.Unmarshal(body, &urls); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	respondingServerAddress := fmt.Sprintf("%v://%v/", scheme, ctx.Request.Host)
	if a.flags.GetRespondingServer() != "" {
		respondingServerAddress = fmt.Sprintf("%v/", a.flags.GetRespondingServer())
	}

	switch a.GetStorageType() {
	case "database":
		db := database.NewInitializerReaderWriter(a.repo, a.flags)
		if err := db.CreateTableIfNotExists(); err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}

		urls, err = db.WriteBatchURL(urls, respondingServerAddress)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusCreated, urls)
	case "file":
		Producer, err := file.NewProducer(a.flags.GetFileStoragePath())
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		defer Producer.Close()

		response := make([]storage.URLStorage, 0, len(urls))
		for _, url := range urls {
			if url.ShortURL == "" {
				url.ShortURL = fmt.Sprintf("%v%v", respondingServerAddress, utils.RandString(10))
			}
			if url.ID == 0 {
				url.ID = uint(rand.Int())
			}
			Producer.WriteEvent(&url)
			response = append(response, storage.URLStorage{CorrelationID: url.CorrelationID, ShortURL: url.ShortURL})
			err = a.repo.WriteShortAndOriginalURL(url.ShortURL, url.OriginalURL)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusCreated, response)
	case "memory":
		response := make([]storage.URLStorage, 0, len(urls))
		for _, url := range urls {
			if url.ShortURL == "" {
				url.ShortURL = fmt.Sprintf("%v%v", respondingServerAddress, utils.RandString(10))
			}
			err = a.repo.WriteShortAndOriginalURL(url.ShortURL, url.OriginalURL)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, errorResponse(err))
				return
			}
			response = append(response, storage.URLStorage{CorrelationID: url.CorrelationID, ShortURL: url.ShortURL})
		}
		ctx.JSON(http.StatusCreated, response)
	}
}

func (a *api) GetStorageType() string {
	if a.flags.GetDatabaseDSN() != "" {
		return "database"
	} else if a.flags.GetFileStoragePath() != "" {
		return "file"
	} else {
		return "memory"
	}
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
