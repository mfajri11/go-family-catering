package migration

import (
	"errors"
	"family-catering/config"
	"family-catering/pkg/consts"
	"family-catering/pkg/logger"
	"fmt"
	"time"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	defaultAttempts = consts.DefaultAttemptsMigration
	defaultTimeout  = consts.DefaultTimeoutMigration
)

var m *migrate.Migrate

func newOrGet() *migrate.Migrate {
	var (
		err      error
		attempts = defaultAttempts
	)
	if m == nil {
		for attempts > 0 {
			m, err = migrate.New("file://migrations", config.Cfg().Postgres.URL())
			if err == nil {
				break
			}
			logger.Info("migration.NewOrGet: Migrate - postgres trying to connect, attempts left %d", attempts)
			time.Sleep(defaultTimeout)
			attempts--

		}

		if err != nil {
			logger.Fatal(err, "migration.NewOrGet: error Migrate connection, err %s", err.Error())

		}
	}

	return m
}

func handleMigrateError(err error) error {
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("migration: %w", err)
	}
	return nil
}

func Up() error {
	m := newOrGet()
	err := m.Up()
	return handleMigrateError(err)
}

func Down() error {
	m := newOrGet()
	err := m.Down()
	return handleMigrateError(err)
}

func Step(n int) error {
	m := newOrGet()
	err := m.Steps(n)
	return handleMigrateError(err)
}

func Drop() error {
	m := newOrGet()
	err := m.Drop()
	return handleMigrateError(err)

}

func Close() (error, error) {
	sourceErr, dbErr := m.Close()
	if sourceErr != nil {
		sourceErr = fmt.Errorf("migration.Close: %w", sourceErr)
		logger.Error(sourceErr, "migration error close source file")
	}

	if dbErr != nil {
		dbErr = fmt.Errorf("migration.Close: %w", dbErr)
		logger.Error(dbErr, "migration error close db")
	}

	if sourceErr == nil && dbErr == nil {
		m = nil
	}

	return sourceErr, dbErr

}
