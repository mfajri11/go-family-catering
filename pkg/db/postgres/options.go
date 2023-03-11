package postgres

import (
	"database/sql"
	"time"
)

type Option func(*sql.DB)

func WithMaxIdleConns(n int) Option {
	return func(db *sql.DB) {
		db.SetMaxIdleConns(n)
	}
}

func WithMaxLifeTime(t time.Duration) Option {
	return func(db *sql.DB) {
		db.SetConnMaxLifetime(t)
	}
}

func WithMaxOpenConnection(n int) Option {
	return func(db *sql.DB) {
		db.SetMaxOpenConns(n)
	}
}
