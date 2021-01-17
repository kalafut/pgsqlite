package pgsqlite

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestSQLite(t *testing.T) {
	assert := assert.New(t)

	dbfile := filepath.Join(os.TempDir(), "test.db")
	os.Remove(dbfile)

	dbRaw, err := sql.Open("sqlite3", dbfile)
	assert.Nil(err)

	db := New(dbRaw)

	_, err = db.Exec(`CREATE TABLE users(id INTEGER, first TEXT, last, TEXT, age INTEGER);`)
	assert.Nil(err)

	_, err = db.Exec(`INSERT INTO users(first, last) VALUES('John', 'Doe');`)
	assert.Nil(err)

	row := db.QueryRow(
		`SELECT first, last FROM users WHERE last=$2 AND first=$1`, "John", "Doe")
	assert.Nil(row.Err())

	var first, last string
	assert.Nil(row.Scan(&first, &last))
	assert.Equal("John", first)
	assert.Equal("Doe", last)

	assert.Nil(db.Close())

	// Lets try Open
	db, err = Open("sqlite3", dbfile)
	assert.Nil(err)

	row = db.QueryRow(`INSERT INTO users(first, last) VALUES('Jane', 'Smith') RETURNING id;`)
	assert.Nil(row.Err())

	var id int
	assert.Nil(row.Scan(&id))
	assert.Equal(2, id)

	row = db.QueryRow(
		`SELECT first, last FROM users WHERE last=$2 AND first=$1`, "Jane", "Smith")
	assert.Nil(row.Err())

	assert.Nil(row.Scan(&first, &last))
	assert.Equal("Jane", first)
	assert.Equal("Smith", last)
}

func Test_rewrite(t *testing.T) {
	tests := map[string]struct {
		in  string
		out string
	}{

		"selectjbasic": {
			in:  "SELECT * FROM USERS;",
			out: "SELECT * FROM USERS;",
		},
		"select params": {
			in:  "SELECT * FROM USERS WHERE age > $2 AND height < $1;",
			out: "SELECT * FROM USERS WHERE age > ?2 AND height < ?1;",
		},
		"insert params": {
			in:  "INSERT INTO USERS (a,b,c) VALUES ($1, $2, $3)",
			out: "INSERT INTO USERS (a,b,c) VALUES (?1, ?2, ?3)",
		},
		"insert returning": {
			in:  "INSERT INTO USERS (a,b,c) VALUES ($1, $2, $3) RETURNING a",
			out: "INSERT INTO USERS (a,b,c) VALUES (?1, ?2, ?3)",
		},
		"insert lower case returning": {
			in:  "insert into users (a,b,c) values ($1, $2, $3) returning a",
			out: "insert into users (a,b,c) values (?1, ?2, ?3)",
		},
		"insert returning ;": {
			in:  "insert into users (a,b,c) values ($1, $2, $3) returning a;",
			out: "insert into users (a,b,c) values (?1, ?2, ?3)",
		},
		"insert returning * (no change)": {
			in:  "insert into users (a,b,c) values ($1, $2, $3) returning *;",
			out: "insert into users (a,b,c) values (?1, ?2, ?3) returning *;",
		},
		"insert returning multiple (no change)": {
			in:  "insert into users (a,b,c) values ($1, $2, $3) returning a, b",
			out: "insert into users (a,b,c) values (?1, ?2, ?3) returning a, b",
		},
	}
	for name, test := range tests {
		act, _ := rewrite(test.in)
		if act != test.out {
			t.Errorf("error in %q. Expected %q, got %q", name, test.out, act)
		}
	}
}
