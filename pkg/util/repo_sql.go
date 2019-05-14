package util

import (
	"database/sql"
	"fmt"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"strings"
)

var (
	countStmtTemplate     string
	deleteAllStmtTemplate string
	listStmtTemplate      string
	fetchStmtTemplate     string
	createStmtTemplate    string
	updateStmtTemplate    string
	deleteOneStmtTemplate string
)

func init() {
	countStmtTemplate = "SELECT COUNT(*) FROM %s WHERE deleted = 0"
	deleteAllStmtTemplate = "DELETE FROM %s"
	listStmtTemplate = "SELECT id, version, organisation, attributes FROM %s  WHERE deleted = 0 LIMIT $1 OFFSET $2"
	fetchStmtTemplate = "SELECT id, version, organisation, attributes FROM %s WHERE id = $1 AND deleted = 0"
	createStmtTemplate = "INSERT INTO %s (id, version, organisation, attributes) VALUES ($1, $2, $3, $4)"
	updateStmtTemplate = "UPDATE %s SET attributes=$1, version=$2 WHERE id=$3 AND version=$4"
	deleteOneStmtTemplate = "UPDATE %s SET deleted=1 WHERE id=$1 AND version=$2"
}

type SqlRepo struct {
	db            *sql.DB
	schema        string
	countStmt     string
	deleteAllStmt string
	listStmt      string
	fetchStmt     string
	createStmt    string
	updateStmt    string
	deleteOneStmt string
}

func (repo *SqlRepo) fmtTemplate(tpl string) string {
	return fmt.Sprintf(tpl, repo.schema)
}

func (repo *SqlRepo) Init() error {
	if repo.schema == "" {
		return fmt.Errorf("no schema defined")
	}

	repo.countStmt = repo.fmtTemplate(countStmtTemplate)
	repo.deleteAllStmt = repo.fmtTemplate(deleteAllStmtTemplate)
	repo.listStmt = repo.fmtTemplate(listStmtTemplate)
	repo.fetchStmt = repo.fmtTemplate(fetchStmtTemplate)
	repo.createStmt = repo.fmtTemplate(createStmtTemplate)
	repo.updateStmt = repo.fmtTemplate(updateStmtTemplate)
	repo.deleteOneStmt = repo.fmtTemplate(deleteOneStmtTemplate)
	return nil
}

func (repo *SqlRepo) Close() error {
	return repo.db.Close()
}

func (repo *SqlRepo) Check() error {
	return repo.db.Ping()
}

func (repo *SqlRepo) List(offset int, limit int) ([]*RepoItem, error) {
	items := []*RepoItem{}

	rows, err := repo.db.Query(repo.listStmt, limit, offset)
	if err != nil {
		return items, errors.Wrap(err, repo.listStmt)
	}

	defer rows.Close()

	for rows.Next() {
		item := &RepoItem{}
		err := rows.Scan(&item.Id, &item.Version, &item.Organisation, &item.Attributes)
		if err != nil {
			return items, errors.Wrap(err, "Error parsing database row")
		}
		items = append(items, item)

	}
	return items, nil
}

func (repo *SqlRepo) Fetch(item *RepoItem) (*RepoItem, error) {
	found := &RepoItem{}

	rows, err := repo.db.Query(repo.fetchStmt, item.Id)
	if err != nil {
		return found, errors.Wrap(err, repo.fetchStmt)
	}

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&found.Id, &found.Version, &found.Organisation, &found.Attributes)
		if err != nil {
			return found, errors.Wrap(err, "Error parsing database row")
		}

		return found, nil

	}

	return found, fmt.Errorf("DB_NOT_FOUND")
}

func (repo *SqlRepo) Create(item *RepoItem) (*RepoItem, error) {
	stmt, err := repo.db.Prepare(repo.createStmt)
	if err != nil {
		return item, errors.Wrap(err, repo.createStmt)
	}

	defer stmt.Close()
	_, err = stmt.Exec(item.Id, 0, item.Organisation, item.Attributes)
	if err != nil {
		errorCode := "DB_ERROR"
		if strings.Contains(strings.ToLower(err.Error()), "unique constraint") {
			errorCode = "DB_CONFLICT"
		}
		return item, errors.Wrap(err, errorCode)
	}

	item.Version = 0
	return item, nil
}

func (repo *SqlRepo) Update(item *RepoItem) (*RepoItem, error) {
	stmt, err := repo.db.Prepare(repo.updateStmt)
	if err != nil {
		return item, errors.Wrap(err, repo.updateStmt)
	}

	defer stmt.Close()

	newVersion := item.Version + 1

	res, err := stmt.Exec(item.Attributes, newVersion, item.Id, item.Version)
	if err != nil {
		errorCode := "DB_ERROR"
		return item, errors.Wrap(err, errorCode)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		errorCode := "DB_ERROR"
		return item, errors.Wrap(err, errorCode)
	}

	switch rowsAffected {
	case 0:
		return item, errors.New("DB_CONFLICT")
	case 1:
		item.Version = newVersion
		return item, nil
	default:
		return item, fmt.Errorf("DB_ERROR: more than 1 row affected by update: %v", rowsAffected)
	}
}

func (repo *SqlRepo) Delete(item *RepoItem) error {
	stmt, err := repo.db.Prepare(repo.deleteOneStmt)
	if err != nil {
		return errors.Wrap(err, repo.deleteOneStmt)
	}

	defer stmt.Close()
	res, err := stmt.Exec(item.Id, item.Version)
	if err != nil {
		errorCode := "DB_ERROR"
		return errors.Wrap(err, errorCode)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		errorCode := "DB_ERROR"
		return errors.Wrap(err, errorCode)
	}

	switch rowsAffected {
	case 0:
		return errors.New("DB_NOT_FOUND")
	case 1:
		return nil
	default:
		return fmt.Errorf("DB_ERROR: more than 1 row affected by delete: %v", rowsAffected)
	}
}

func (repo *SqlRepo) IsConflict(err error) bool {
	return strings.Contains(err.Error(), "DB_CONFLICT")
}

func (repo *SqlRepo) IsNotFound(err error) bool {
	return strings.Contains(err.Error(), "DB_NOT_FOUND")
}

func (repo *SqlRepo) DeleteAll() error {
	stmt, err := repo.db.Prepare(repo.deleteAllStmt)
	if err != nil {
		return errors.Wrap(err, repo.deleteAllStmt)
	}

	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		errorCode := "DB_ERROR"
		// TODO: better translate errors
		return errors.Wrap(err, errorCode)
	}

	return nil
}

func (repo *SqlRepo) Info() (RepoInfo, error) {
	var count int
	var info RepoInfo

	rows, err := repo.db.Query(repo.countStmt)
	if err != nil {
		return info, errors.Wrap(err, repo.countStmt)
	}

	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return info, errors.Wrap(err, repo.countStmt)
		}
		break
	}

	return RepoInfo{Count: count}, nil
}
