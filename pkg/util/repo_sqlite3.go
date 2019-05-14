package util

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type Sqlite3Repo struct {
	SqlRepo
	backend string
}

func NewSqlite3Repo(config RepoConfig) (Repo, error) {
	backend := config.Uri
	if backend == "" {
		backend = "file::memory:?cache=shared"
	}

	repo := &Sqlite3Repo{
		SqlRepo: SqlRepo{
			schema: config.Schema,
		},
		backend: backend,
	}

	database, err := sql.Open("sqlite3", backend)
	if err != nil {
		return repo, errors.Wrap(err, "Unable to connect to the database")
	}

	if config.Migrations != "" {
		driver, err := sqlite3.WithInstance(database, &sqlite3.Config{})
		if err != nil {
			return repo, errors.Wrap(err, "Could not start migration")
		}

		m, err := migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%s", config.Migrations),
			"sqlite3", driver)

		if err != nil {
			return repo, errors.Wrap(err, "Migration failed")
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return repo, errors.Wrap(err, "Error while syncing")
		}
	}

	repo.db = database
	return repo, nil
}

func (repo *Sqlite3Repo) Description() string {
	return fmt.Sprintf("sqlite3 (%s)", repo.backend)
}
