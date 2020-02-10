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
	port          string
	host          string
)

//Postgres wrapper struct, should be easy to swap out if we want different db connections
type Postgres struct {
	db *sql.DB
}

//NewClient creates a new postgres database client
func NewClient() (Postgres, error) {

	dbURL, envVariableOk := os.LookupEnv("DATABASE_URL")
	if !envVariableOk {
		fmt.Println("DATABASE_URL not set.")
	}

	psqlConnection, err := sql.Open("postgres", dbURL)
	if err != nil {
		return Postgres{db: nil}, err
	}

	err = psqlConnection.Ping()
	if err != nil {
		return Postgres{db: nil}, err
	}

	return Postgres{db: psqlConnection}, nil
}

//FindRedirectURLByID returns the record in the database with the given ID
func (p Postgres) FindRedirectURLByID(linkID byte) (string, error) {

	var result string

	sqlStatement := `SELECT url FROM links WHERE id=$1;`

	row := p.db.QueryRow(sqlStatement, linkID)
	err := row.Scan(&result)

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

	_, statisticsErr := p.db.Exec(statisticsSQL, linkID)

	return statisticsErr
}

//InsertURL actually does the db insert when creating a shortened link
func (p Postgres) InsertURL(link string) (int, error) {
	var id int

	sqlStatement := `INSERT INTO links (url)
					 VALUES ($1)
					 RETURNING id`

	queryErr := p.db.QueryRow(sqlStatement, link).Scan(&id)

	return id, queryErr
}

//GetLinkViewCount queries the view data for total number of times a link has been viewed
func (p Postgres) GetLinkViewCount(id int) (int, error) {
	var count int

	sqlStatement := `SELECT COUNT(*) FROM link_statistics WHERE link_id=$1;`

	row := p.db.QueryRow(sqlStatement, id)
	err := row.Scan(&count)

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
