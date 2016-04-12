package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type GspkDb struct {
	*sql.DB
}

func DbConn(host, user, password string) (GspkDb, error) {
	connstr := fmt.Sprintf("host=%s user=%s password=%s sslmode=disable", host, user, password)

	db, err := sql.Open("postgres", connstr)
	if err != nil {
		return GspkDb{nil}, err
	}

	return GspkDb{db}, nil
}
