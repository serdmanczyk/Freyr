package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func DbConn(host, user, password string) (DB, error) {
	connstr := fmt.Sprintf("host=%s user=%s password=%s sslmode=disable", host, user, password)

	db, err := sql.Open("postgres", connstr)
	if err != nil {
		return DB{nil}, err
	}

	return DB{db}, nil
}
