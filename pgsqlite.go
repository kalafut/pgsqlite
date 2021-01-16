package pgsqlite

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"

	"github.com/kalafut/q"
)

type MyDriver struct {
	*sql.DB
	//sqlite bool
}

type result struct {
	id int64
}

func (r *result) LastInsertId() (int64, error) {
	return r.id, nil
}

func (r *result) RowsAffected() (int64, error) {
	return 0, errors.New("RowsAffected is not supported by this wrapper")
}

func (db *MyDriver) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.QueryRowContext(context.Background(), query, args...)
}

func (db *MyDriver) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	query = strings.TrimSpace(query)

	re2 := regexp.MustCompile(`RETURNING\s+\w+`)

	fetchID := re2.FindString(query) != ""

	query = re2.ReplaceAllString(query, "")
	query = rebind(query)

	q.Q(query)
	if fetchID {
		q.Q("here")
		db.DB.ExecContext(ctx, query, args...)
		//if row.Err() != nil {
		//	return row
		//}

		return db.DB.QueryRow("SELECT last_insert_rowid();")
	}

	//q.Q(query)
	//r, err := db.DB.ExecContext(ctx, query, args...)
	//id, err := r.LastInsertId()
	//q.Q(id, err)

	//row := db.DB.QueryRowContext(ctx, query, args...)
	//q.Q(row.)

	return db.DB.QueryRowContext(ctx, query, args...)
}

func rewrite(query string) string {
	re2 := regexp.MustCompile(`(?i)\s*RETURNING\s+\w+;?\s*$`)
	q.Q(re2.FindString(query))

	query = re2.ReplaceAllString(query, "")
	query = rebind(query)

	return query
}

//func (db *MyDriver) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
//	var err error
//	query = strings.TrimSpace(query)
//
//	re := regexp.MustCompile(`^INSERT.*RETURNING\s+\w+`)
//	re2 := regexp.MustCompile(`RETURNING\s+\w+`)
//
//	if db.sqlite {
//		query = re2.ReplaceAllString(query, "")
//		query, err = normalizeQuery(query)
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	if re.MatchString(query) {
//		r := &result{}
//		rows := db.QueryRowContext(ctx, query, args...)
//		err := rows.Scan(&r.id)
//
//		return r, err
//	}
//
//	return db.DB.ExecContext(ctx, query, args...)
//}

func Open(driverName, dataSourceName string) (*MyDriver, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return &MyDriver{
		DB: db,
		//sqlite: strings.TrimSpace(driverName) == "sqlite3",
	}, nil
}

var paramRe = regexp.MustCompile(`\$\d+`)

// rebind will replace postgres $1, $2, ... parameters with
// with sqlite's ?1, ?2 style.
func rebind(s string) string {
	s = paramRe.ReplaceAllStringFunc(s, func(m string) string {
		return "?" + m[1:]
	})
	return s
}

//func main() {
//	d, err := Open("pgx", `postgres://kalafut@localhost/snipstream?sslmode=disable`)
//	if err != nil {
//		panic(err)
//	}
//
//	result, err := d.Exec("INSERT INTO snippets(text) VALUES($1) RETURNING id", "abc")
//	//result, err := d.Exec("INSERT INTO snippets(text) VALUES($1)", "abc")
//	if err != nil {
//		panic(err)
//	}
//
//	fmt.Println(result.RowsAffected())
//	fmt.Println(result.LastInsertId())
//
//	db, err := Open("sqlite3", "./foo.db")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	result, err = db.Exec("INSERT INTO snippets(text) VALUES($1) RETURNING id", "abc")
//	//result, err := d.Exec("INSERT INTO snippets(text) VALUES($1)", "abc")
//	if err != nil {
//		panic(err)
//	}
//
//	fmt.Println(result.RowsAffected())
//	fmt.Println(result.LastInsertId())
//
//	_, err = d.Query("SELECT * from snippets")
//	if err != nil {
//		panic(err)
//	}
//
//}
//
