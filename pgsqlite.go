package pgsqlite

import (
	"context"
	"database/sql"
	"regexp"
)

var paramRe = regexp.MustCompile(`\$\d+`)
var returnRe = regexp.MustCompile(`(?i)\s*RETURNING\s+\w+;?\s*$`)

type MyDriver struct {
	*sql.DB
}

func Open(driverName, dataSourceName string) (*MyDriver, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return New(db), nil
}

func New(db *sql.DB) *MyDriver {
	return &MyDriver{
		DB: db,
	}
}

func (db *MyDriver) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.QueryRowContext(context.Background(), query, args...)
}

func (db *MyDriver) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	query, returnStmt := rewrite(query)

	if returnStmt {
		result, err := db.DB.ExecContext(ctx, query, args...)
		if err != nil {
			return nil
		}
		id, err := result.LastInsertId()
		if err != nil {
			return nil
		}

		return db.DB.QueryRow("SELECT ?", id)
	}

	return db.DB.QueryRowContext(ctx, query, args...)
}

func rewrite(query string) (string, bool) {
	// replace RETURN statements
	queryNew := returnRe.ReplaceAllString(query, "")

	returnStmt := (queryNew != query)

	// replace postgres-sql $1, $2, ... parameters with
	// with sqlite's ?1, ?2 style.
	queryNew = paramRe.ReplaceAllStringFunc(queryNew, func(m string) string {
		return "?" + m[1:]
	})

	return queryNew, returnStmt
}
