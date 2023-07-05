package storage

import (
	"errors"

	"github.com/volnistii11/URL-shortener/internal/app/utils"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	ReadURL(id string) (string, error)
	WriteURL(url string) (string, error)
	WriteShortAndOriginalURL(shortURL, originalURL string) error
	SetRestoreData(shortURL string, originalURL string)
	GetDatabase() *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &url{
		urlDependency: map[string]string{},
		db:            db,
	}
}

type url struct {
	urlDependency map[string]string
	db            *sqlx.DB
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

func (storage *url) WriteShortAndOriginalURL(shortURL, originalURL string) error {
	storage.urlDependency[shortURL] = originalURL
	if len(storage.urlDependency[shortURL]) < 10 {
		return errors.New("error in short link generation, link length is less than 10")
	}
	return nil
}

func (storage *url) SetRestoreData(shortURL string, originalURL string) {
	storage.urlDependency[shortURL] = originalURL
}

func (storage *url) GetDatabase() *sqlx.DB {
	return storage.db
}
