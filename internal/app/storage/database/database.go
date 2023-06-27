package database

import (
	"fmt"
	"github.com/volnistii11/URL-shortener/internal/model"

	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	"github.com/volnistii11/URL-shortener/internal/app/utils"

	"github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pkg/errors"
)

type InitializerReaderWriter interface {
	CreateTableIfNotExists() error
	ReadURL(shortURL string) (string, error)
	WriteURL(urls *model.URL) (string, error)
	WriteBatchURL(urls []model.URL, serverAddress string) ([]model.URL, error)
	ReadBatchURLByUserID(userId int) ([]model.URL, error)
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

func (db *database) CreateTableIfNotExists() error {
	if err := db.repo.GetDatabase().Ping(); err != nil {
		return err
	}

	//if err := runMigrations(db.flags.GetDatabaseDSN()); err != nil {
	//	return errors.Wrap(err, "Start migrations")
	//}

	_, err := db.repo.GetDatabase().
		Exec("CREATE TABLE IF NOT EXISTS url_dependencies (id serial primary key, correlation_id varchar(255) null, short_url varchar(255) not null unique, original_url varchar(255) not null unique, user_id integer null)")
	if err != nil {
		return err
	}

	return nil
}

func (db *database) ReadURL(shortURL string) (string, error) {
	dbConnection := db.repo.GetDatabase()
	if err := dbConnection.Ping(); err != nil {
		return "", err
	}

	sb := squirrel.Select("original_url").
		From("url_dependencies").
		Where(squirrel.Eq{"short_url": shortURL}).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(dbConnection)

	var originalURL string
	err := sb.QueryRow().Scan(&originalURL)
	if err != nil {
		return "", err
	}

	return originalURL, nil
}

func (db *database) WriteURL(url *model.URL) (string, error) {
	dbConnection := db.repo.GetDatabase()

	if err := dbConnection.Ping(); err != nil {
		return "", err
	}

	if url.ShortURL == "" {
		url.ShortURL = utils.RandString(10)
	}

	sb := squirrel.StatementBuilder.
		Insert("url_dependencies").
		Columns("correlation_id", "short_url", "original_url", "user_id").
		PlaceholderFormat(squirrel.Dollar).
		RunWith(dbConnection)

	sb = sb.Values(
		url.CorrelationID,
		url.ShortURL,
		url.OriginalURL,
		url.UserID,
	)

	_, err := sb.Exec()
	if err != nil {
		var shortURL string
		sb := squirrel.Select("short_url").
			From("url_dependencies").
			Where(squirrel.Eq{"original_url": url.OriginalURL}).
			PlaceholderFormat(squirrel.Dollar).
			RunWith(dbConnection)

		errSelect := sb.QueryRow().Scan(&shortURL)
		if errSelect != nil {
			return "", errors.Wrap(errSelect, "Select")
		}
		return shortURL, err
	}

	return url.ShortURL, nil
}

func (db *database) WriteBatchURL(urls []model.URL, serverAddress string) ([]model.URL, error) {
	if err := db.repo.GetDatabase().Ping(); err != nil {
		return nil, err
	}

	tx, err := db.repo.GetDatabase().Begin()
	if err != nil {
		return nil, err
	}

	response := make([]model.URL, 0, len(urls))

	sb := squirrel.StatementBuilder.
		Insert("url_dependencies").
		Columns("correlation_id", "short_url", "original_url", "user_id").
		PlaceholderFormat(squirrel.Dollar).
		RunWith(tx)

	for _, url := range urls {
		if url.ShortURL == "" {
			url.ShortURL = utils.RandString(10)
		}

		sb = sb.Values(
			url.CorrelationID,
			url.ShortURL,
			url.OriginalURL,
			url.UserID,
		)

		shortURL := fmt.Sprintf("%v%v", serverAddress, url.ShortURL)
		response = append(response, model.URL{CorrelationID: url.CorrelationID, ShortURL: shortURL})
	}

	_, err = sb.Exec()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, errors.Wrap(err, "Rollback")
		}
		return nil, errors.Wrap(err, "Query")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "Commit")
	}

	return response, nil
}

func (db *database) ReadBatchURLByUserID(userID int) ([]model.URL, error) {
	if err := db.repo.GetDatabase().Ping(); err != nil {
		return nil, err
	}

	tx, err := db.repo.GetDatabase().Begin()
	if err != nil {
		return nil, err
	}

	queryRowCount := squirrel.Select("COUNT(*)").
		From("url_dependencies").
		Where(squirrel.Eq{"user_id": userID}).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(tx)

	var rowCount int
	errSelect := queryRowCount.QueryRow().Scan(&rowCount)
	if errSelect != nil {
		if err := tx.Rollback(); err != nil {
			return nil, errors.Wrap(err, "Select row count -> rollback")
		}
		return nil, errors.Wrap(errSelect, "Select row count")
	}
	if rowCount == 0 {
		if err := tx.Rollback(); err != nil {
			return nil, errors.Wrap(err, "Row count = 0 -> rollback")
		}
		return nil, errors.Wrap(err, "Row count = 0")
	}

	query := squirrel.Select("short_url, original_url").
		From("url_dependencies").
		Where(squirrel.Eq{"user_id": userID}).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(tx)
	rows, errSelect := query.Query()
	if errSelect != nil {
		if err := tx.Rollback(); err != nil {
			return nil, errors.Wrap(err, "Select urls -> rollback")
		}
		return nil, errors.Wrap(err, "Select urls")
	}
	defer rows.Close()

	response := make([]model.URL, 0, rowCount)
	var shortURL string
	var originalURL string
	for rows.Next() {
		err = rows.Scan(&shortURL, &originalURL)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				return nil, errors.Wrap(err, "Scan -> rollback")
			}
			return nil, errors.Wrap(err, "Scan")
		}
		response = append(response, model.URL{ShortURL: shortURL, OriginalURL: originalURL})
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "Commit")
	}

	return response, nil
}

func runMigrations(dsn string) error {
	const migrationsPath = "../../internal/app/storage/database/migrations"

	m, err := migrate.New(fmt.Sprintf("file://%s", migrationsPath), dsn)
	if err != nil {
		return errors.Wrap(err, "Create migrations")
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return errors.Wrap(err, "Run migrations")
		}
	}
	return nil
}
