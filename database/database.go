// Package database defines method for interacting with PostgreSQL for
// persistent data.
package database

import (
	"database/sql"
	"fmt"
)

// DB is a type used by external packages to implement database operations.
// DB should be initialized via DBConn function.
type DB struct {
	*sql.DB
}

// DBConn returns a new DB, initialized to connect to the database
func DBConn(dbType, host, user, password string) (DB, error) {
	connstr := fmt.Sprintf("host=%s user=%s password=%s sslmode=disable", host, user, password)

	db, err := sql.Open(dbType, connstr)
	if err != nil {
		return DB{nil}, err
	}

	return DB{db}, nil
}
