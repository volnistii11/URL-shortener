package storage

import (
	"errors"
	"github.com/volnistii11/URL-shortener/internal/app/utils"
)

type Repository interface {
	ReadURL(id string) (string, error)
	WriteURL(url string) (string, error)
	SetRestoreData(shortURL string, originalURL string)
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
	fullURL, ok := storage.urlDependency[id]
	if !ok {
		return fullURL, errors.New("full url not found")
	}
	return fullURL, nil
}

func (storage *url) WriteURL(url string) (string, error) {
	shortURL := utils.RandString(10)
	storage.urlDependency[shortURL] = url
	if len(storage.urlDependency[shortURL]) < 10 {
		return shortURL, errors.New("error in short link generation, link length is less than 10")
	}
	return shortURL, nil
}

func (storage *url) SetRestoreData(shortURL string, originalURL string) {
	storage.urlDependency[shortURL] = originalURL
}
