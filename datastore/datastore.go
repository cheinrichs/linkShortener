package datastore

import (
	"database/sql"

	postgres "github.com/cheinrichs/linkShortener/datastore/postgres"

	_ "github.com/lib/pq"
)

var (
	envVariableOk bool
	dbURL         string
	port          string
	host          string
	db            *sql.DB
)

//DBClient is used to make calls to the database.
type DBClient interface {
	findRedirectURLByID(linkID byte) (string, error)

	recordView(linkID byte) error

	insertURL(link string) (int, error)

	getLinkViewCount(id int) (int, error)

	initializeEnv()
}

//NewClient initializes a DB client
func NewClient() (DBClient, error) {
	return postgres.NewClient()
}
