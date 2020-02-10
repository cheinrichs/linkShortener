package datastore

import (
	"database/sql"
	"fmt"
	"os"

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

//Postgres contains all postgresql implementations for the DBClient interface
type Postgres struct {
}

//NewClient creates a new postgres database client
func (p Postgres) NewClient() (*sql.DB, error) {

	dbURL, envVariableOk = os.LookupEnv("DATABASE_URL")
	if !envVariableOk {
		fmt.Println("DATABASE_URL not set.")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

//FindRedirectURLByID returns the record in the database with the given ID
func (p Postgres) FindRedirectURLByID(linkID byte) (string, error) {

	var result string

	sqlStatement := `SELECT url FROM links WHERE id=$1;`

	row := db.QueryRow(sqlStatement, linkID)
	err := row.Scan(&result)

	db.Close()

	switch err {
	case sql.ErrNoRows:
		return "", nil
	case nil:
		return result, nil
	default:
		return "", err
	}
}

//RecordView increments the view statistics by adding a record to the link_statistics table
func (p Postgres) RecordView(linkID byte) error {

	statisticsSQL := `INSERT INTO link_statistics (link_id)
					 VALUES ($1)`

	_, statisticsErr := db.Exec(statisticsSQL, linkID)

	db.Close()
	return statisticsErr
}

//InsertURL actually does the db insert when creating a shortened link
func (p Postgres) InsertURL(link string) (int, error) {
	var id int

	sqlStatement := `INSERT INTO links (url)
					 VALUES ($1)
					 RETURNING id`

	queryErr := db.QueryRow(sqlStatement, link).Scan(&id)

	db.Close()
	return id, queryErr
}

//GetLinkViewCount queries the view data for total number of times a link has been viewed
func (p Postgres) GetLinkViewCount(id int) (int, error) {
	var count int

	sqlStatement := `SELECT COUNT(*) FROM link_statistics WHERE link_id=$1;`

	row := db.QueryRow(sqlStatement, id)
	err := row.Scan(&count)

	db.Close()

	switch err {
	case sql.ErrNoRows:
		count = 0
		return count, err
	case nil:
		return count, err
	default:
		return -1, err
	}
}
