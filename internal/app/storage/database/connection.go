package database

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func NewConnection(driver string, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open(driver, dsn)
	if err != nil {
		errors.Wrap(err, "failed to create a db connection")
		return nil, err
	}

	if err = db.Ping(); err != nil {
		errors.Wrap(err, "failed to ping the db")
		return nil, err
	}

	return db, nil
}
