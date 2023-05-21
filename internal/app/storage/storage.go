package storage

import (
	"errors"
	"github.com/volnistii11/URL-shortener/internal/app/utils"
)

type URL struct {
	URLDependency map[string]string
}

func NewStorage() *URL {
	return &URL{
		URLDependency: map[string]string{},
	}
}

func (storage *URL) ReadURL(id string) (string, error) {
	if fullURL, ok := storage.URLDependency[id]; !ok {
		return fullURL, errors.New("full url not found")
	} else {
		return fullURL, nil
	}
}

func (storage *URL) WriteURL(url string) string {
	shortURL := utils.RandString(10)
	storage.URLDependency[shortURL] = url
	return shortURL
}

type URLReaderWriter interface {
	ReadURL(id string) (string, error)
	WriteURL(url string) string
}

var myURL URLReaderWriter = NewStorage()

func GetStorage() URLReaderWriter {
	return myURL
}
