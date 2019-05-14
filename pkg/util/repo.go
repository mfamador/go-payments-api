package util

import (
	"fmt"
)

type RepoItem struct {
	Id           string `db:"id"`
	Version      int    `db:"version"`
	Organisation string `db:"organisation"`
	Attributes   string `db:"attributes"`
}

type RepoInfo struct {
	Count int `json:"count"`
}

type RepoConfig struct {
	Driver     string
	Uri        string
	Migrations string
	Schema     string
}

type Repo interface {
	Init() error
	Description() string
	Info() (RepoInfo, error)
	Check() error
	Close() error
	List(offset int, limit int) ([]*RepoItem, error)
	Create(item *RepoItem) (*RepoItem, error)
	Update(item *RepoItem) (*RepoItem, error)
	Fetch(item *RepoItem) (*RepoItem, error)
	Delete(item *RepoItem) error
	DeleteAll() error
	IsConflict(err error) bool
	IsNotFound(err error) bool
}

func NewRepo(config RepoConfig) (Repo, error) {
	var db Repo
	switch config.Driver {
	case "sqlite3":
		return NewSqlite3Repo(config)
	case "postgres":
		return NewPostgresRepo(config)
	default:
		return db, fmt.Errorf("repo driver not supported: %v", config.Driver)
	}
}
