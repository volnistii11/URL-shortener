package database

import (
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	"github.com/volnistii11/URL-shortener/internal/app/utils"
)

type InitializerReaderWriter interface {
	CreateTableIfNotExists() error
	ReadURL(shortURL string) (string, error)
	WriteURL(urls *RequestSchema) (string, error)
}

func NewInitializerReaderWriter(repository storage.Repository, cfg config.Flags) InitializerReaderWriter {
	return &database{
		repo:  repository,
		flags: cfg,
	}
}

type database struct {
	repo  storage.Repository
	flags config.Flags
}

type RequestSchema struct {
	ID          uint   `json:"uuid,string"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (db *database) CreateTableIfNotExists() error {
	if err := db.repo.GetDatabase().Ping(); err != nil {
		return err
	}

	_, err := db.repo.GetDatabase().
		Exec("CREATE TABLE IF NOT EXISTS url_dependencies (id serial primary key, short_url varchar(255) not null unique, original_url varchar(255) not null unique)")
	if err != nil {
		return err
	}

	return nil
}

func (db *database) ReadURL(shortURL string) (string, error) {
	if err := db.repo.GetDatabase().Ping(); err != nil {
		return "", err
	}

	var originalURL string
	err := db.repo.GetDatabase().QueryRow("SELECT original_url FROM url_dependencies WHERE short_url = $1", shortURL).Scan(&originalURL)
	if err != nil {
		return "", err
	}

	return originalURL, nil
}

func (db *database) WriteURL(urls *RequestSchema) (string, error) {
	if err := db.repo.GetDatabase().Ping(); err != nil {
		return "", err
	}

	if urls.ShortURL == "" {
		urls.ShortURL = utils.RandString(10)
	}

	_, err := db.repo.GetDatabase().Exec("INSERT INTO url_dependencies (short_url, original_url) VALUES ($1, $2)", urls.ShortURL, urls.OriginalURL)
	if err != nil {
		return "", err
	}

	return urls.ShortURL, nil
}
