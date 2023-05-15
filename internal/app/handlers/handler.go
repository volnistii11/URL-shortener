package handlers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	"github.com/volnistii11/URL-shortener/internal/app/utils"
	"net/http"
)

func init() {
	storage.URLDependency = map[string]string{}
}

func CreateShortURL(ctx *gin.Context) {
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
	shortURL := utils.RandString(10)
	storage.URLDependency[shortURL] = string(body)

	respondingServerAddress := scheme + "://" + ctx.Request.Host + ctx.Request.RequestURI
	if config.Addresses.RespondingServer != "" {
		respondingServerAddress = config.Addresses.RespondingServer + "/"
	}

	fmt.Println(respondingServerAddress)
	ctx.String(http.StatusCreated, "%v%v", respondingServerAddress, shortURL)
}

func GetFullURL(ctx *gin.Context) {
	shortURL := ctx.Params.ByName("short_url")

	if fullURL, ok := storage.URLDependency[shortURL]; ok {
		ctx.Header("Location", fullURL)
		ctx.Status(http.StatusTemporaryRedirect)
	} else {
		ctx.Status(http.StatusBadRequest)
	}
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
