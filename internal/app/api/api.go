package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	"net/http"
)

type Provider interface {
	CreateShortURL(ctx *gin.Context)
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
	shortURL, err := a.repo.WriteURL(string(body))
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

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
