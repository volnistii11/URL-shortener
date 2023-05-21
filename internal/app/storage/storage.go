package storage

import (
	"errors"
	"github.com/volnistii11/URL-shortener/internal/app/utils"
)

type Repository interface {
	ReadURL(id string) (string, error)
	WriteURL(url string) string
}

func NewRepository() Repository {
	return &url{
		urlDependency: map[string]string{},
	}
}

type url struct {
	urlDependency map[string]string
}

func (storage *url) ReadURL(id string) (string, error) {
	if fullURL, ok := storage.urlDependency[id]; !ok {
		return fullURL, errors.New("full url not found")
	} else {
		return fullURL, nil
	}
}

func (storage *url) WriteURL(url string) string {
	shortURL := utils.RandString(10)
	storage.urlDependency[shortURL] = url
	return shortURL
}
