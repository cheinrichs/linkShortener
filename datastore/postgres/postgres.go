package postgres

import (
	"database/sql"
	"fmt"
	"os"
)

var dbURL string
var envVariableOk bool

//Wrapper contains a pointer to a postgresql db client
type Wrapper struct {
	client *sql.DB
}

//NewClient creates a new postgres database client
func NewClient() (*Wrapper, error) {

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

	return &Wrapper{client: db}, nil
}

//FindRedirectURLByID returns the record in the database with the given ID
func (wrapper *Wrapper) FindRedirectURLByID(linkID byte) (string, error) {
	var result string

	sqlStatement := `SELECT url FROM links WHERE id=$1;`

	row := wrapper.client.QueryRow(sqlStatement, linkID)
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
func (wrapper *Wrapper) RecordView(linkID byte) error {

	statisticsSQL := `INSERT INTO link_statistics (link_id)
					 VALUES ($1)`

	_, statisticsErr := wrapper.client.Exec(statisticsSQL, linkID)

	return statisticsErr
}

//InsertURL actually does the db insert when creating a shortened link
func (wrapper *Wrapper) InsertURL(link string) (int, error) {
	var id int

	sqlStatement := `INSERT INTO links (url)
					 VALUES ($1)
					 RETURNING id`

	queryErr := wrapper.client.QueryRow(sqlStatement, link).Scan(&id)

	return id, queryErr
}

//GetLinkViewCount queries the view data for total number of times a link has been viewed
func (wrapper *Wrapper) GetLinkViewCount(id int) (int, error) {
	var count int
	sqlStatement := `SELECT COUNT(*) FROM link_statistics WHERE link_id=$1;`

	row := wrapper.client.QueryRow(sqlStatement, id)
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
