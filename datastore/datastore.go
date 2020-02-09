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

//NewClient creates a new postgres database client
func NewClient() (*sql.DB, error) {

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
func FindRedirectURLByID(linkID byte) (string, error) {

	db, dbErr := NewClient()
	if dbErr != nil {
		return "", dbErr
	}
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
func RecordView(linkID byte) error {

	db, dbErr := NewClient()
	if dbErr != nil {
		return dbErr
	}

	statisticsSQL := `INSERT INTO link_statistics (link_id)
					 VALUES ($1)`

	_, statisticsErr := db.Exec(statisticsSQL, linkID)

	db.Close()
	return statisticsErr
}

//InsertURL actually does the db insert when creating a shortened link
func InsertURL(link string) (int, error) {
	var id int

	db, dbErr := NewClient()
	if dbErr != nil {
		return -1, dbErr
	}

	sqlStatement := `INSERT INTO links (url)
					 VALUES ($1)
					 RETURNING id`

	queryErr := db.QueryRow(sqlStatement, link).Scan(&id)

	db.Close()
	return id, queryErr
}

//GetLinkViewCount queries the view data for total number of times a link has been viewed
func GetLinkViewCount(id int) (int, error) {
	var count int

	db, dbErr := NewClient()
	if dbErr != nil {
		return -1, dbErr
	}

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
