package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresClient interface {
	Closer
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type Closer interface {
	Close() error
}

var _ PostgresClient = (*sql.DB)(nil)

func DataSourcef(username, password, host string, port int, databaseName string) string {
	dataSource := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		username, password, host, port, databaseName)
	return dataSource
}

func New(source string, opts ...Option) (PostgresClient, error) {

	db, err := sql.Open("postgres", source)
	if err != nil {
		err = fmt.Errorf("postgres.New: %w", err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		err = fmt.Errorf("postgres.New: %w", err)
		return nil, err
	}

	for _, opt := range opts {
		opt(db)
	}
	return db, nil
}
