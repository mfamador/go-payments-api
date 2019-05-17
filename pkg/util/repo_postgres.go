package util

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type PosgresRepo struct {
	SqlRepo
	uri string
}

func NewPostgresRepo(config RepoConfig) (Repo, error) {

	repo := &PosgresRepo{
		SqlRepo: SqlRepo{
			schema: config.Schema,
		},
		uri: config.Uri,
	}

	database, err := sql.Open("postgres", repo.uri)
	if err != nil {
		return repo, errors.Wrap(err, "Unable to connect to the database")
	}

	if config.Migrations != "" {
		driver, err := postgres.WithInstance(database, &postgres.Config{})
		if err != nil {
			return repo, errors.Wrap(err, "Could not start migration")
		}

		m, err := migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%s", config.Migrations),
			"postgres", driver)

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

func (repo *PosgresRepo) Description() string {
	return fmt.Sprintf("postgres (%s)", repo.uri)
}
