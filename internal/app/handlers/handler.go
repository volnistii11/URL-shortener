package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	"github.com/volnistii11/URL-shortener/internal/app/utils"
	"net/http"
)

func init() {
	storage.URLMap = map[string]string{}
}

func CreateShortURL(c *gin.Context) {
	c.Header("content-type", "text/plain; charset=utf-8")
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	if len(body) == 0 {
		c.JSON(http.StatusBadRequest, errorResponse("Body is empty"))
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	shortURL := utils.RandString(10)
	storage.URLMap[shortURL] = string(body)

	respondingServerAddress := scheme + "://" + c.Request.Host + c.Request.RequestURI
	if config.Addresses.RespondingServer != "" {
		respondingServerAddress = config.Addresses.RespondingServer + "/"
	}

	fmt.Println(respondingServerAddress)
	c.String(http.StatusCreated, "%v%v", respondingServerAddress, shortURL)
}

func GetFullURL(c *gin.Context) {
	shortURL := c.Params.ByName("short_url")

	if fullURL, ok := storage.URLMap[shortURL]; ok {
		c.Header("Location", fullURL)
		c.Status(http.StatusTemporaryRedirect)
	} else {
		c.Status(http.StatusBadRequest)
	}
}

func errorResponse(err string) gin.H {
	return gin.H{"error": err}
}
