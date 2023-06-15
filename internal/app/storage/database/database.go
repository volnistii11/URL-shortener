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
	WriteBatchURL(urls []RequestSchema) ([]RequestSchema, error)
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
	ID            uint   `json:"uuid,string,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
	ShortURL      string `json:"short_url,omitempty"`
	OriginalURL   string `json:"original_url,omitempty"`
}

func (db *database) CreateTableIfNotExists() error {
	if err := db.repo.GetDatabase().Ping(); err != nil {
		return err
	}

	_, err := db.repo.GetDatabase().
		Exec("CREATE TABLE IF NOT EXISTS url_dependencies (id serial primary key, correlation_id varchar(255) null unique , short_url varchar(255) not null unique, original_url varchar(255) not null unique)")
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

func (db *database) WriteURL(url *RequestSchema) (string, error) {
	if err := db.repo.GetDatabase().Ping(); err != nil {
		return "", err
	}

	if url.ShortURL == "" {
		url.ShortURL = utils.RandString(10)
	}

	_, err := db.repo.GetDatabase().Exec("INSERT INTO url_dependencies (correlation_id, short_url, original_url) VALUES ($1, $2, $3)", url.CorrelationID, url.ShortURL, url.OriginalURL)
	if err != nil {
		return "", err
	}
	return url.ShortURL, nil
}

func (db *database) WriteBatchURL(urls []RequestSchema) ([]RequestSchema, error) {
	if err := db.repo.GetDatabase().Ping(); err != nil {
		return nil, err
	}

	tx, err := db.repo.GetDatabase().Begin()
	if err != nil {
		return nil, err
	}
	response := make([]RequestSchema, 0, len(urls))
	for _, url := range urls {
		if url.ShortURL == "" {
			url.ShortURL = utils.RandString(10)
		}
		_, err := tx.Exec("INSERT INTO url_dependencies (correlation_id, short_url, original_url) VALUES ($1, $2, $3)",
			url.CorrelationID, url.ShortURL, url.OriginalURL)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				return nil, err
			}
			return nil, err
		}
		response = append(response, RequestSchema{CorrelationID: url.CorrelationID, ShortURL: url.ShortURL})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return response, nil
}
