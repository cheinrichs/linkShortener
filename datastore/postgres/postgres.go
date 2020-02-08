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

//findRedirectURLByID returns the record in the database with the given ID
func (client *Wrapper) findRedirectURLByID(linkID byte) (string, error) {
	var result string

	sqlStatement := `SELECT url FROM links WHERE id=$1;`

	row := Wrapper.client.QueryRow(sqlStatement, linkID)
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

func recordView(linkID byte) error {

}

func insertURL(link string) (int, error) {

}

func getLinkViewCount(id int) (int, error) {

}

func initializeEnv() {

}
