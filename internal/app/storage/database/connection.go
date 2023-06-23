package database

import (
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func NewConnection(driver string, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open(driver, dsn)
	if err != nil {
		zap.L().Error("failed to create a db connection", zap.Error(err))
		return nil, err
	}

	return db, nil
}
