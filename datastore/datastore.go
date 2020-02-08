package datastore

import (
	"database/sql"

	"github.com/cheinrichs/linkShortener/datastore/postgres"

	//Using pg client
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
	FindRedirectURLByID(linkID byte) (string, error)

	RecordView(linkID byte) error

	InsertURL(link string) (int, error)

	GetLinkViewCount(id int) (int, error)
}

//NewClient initializes a DB client
func NewClient() (DBClient, error) {
	return postgres.NewClient()
}
