package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"linkShortener/datastore"

	_ "github.com/lib/pq"
)

//Postgres contains all postgresql implementations for the DBClient interface
type Postgres struct {
	db *sql.DB
}

//DBClient is used to make calls to the database.
type DBClient interface {
	FindRedirectURLByID(linkID byte) (string, error)

	RecordView(linkID byte) error

	InsertURL(link string) (int, error)

	GetLinkViewCount(id int) (int, error)
}

var dbURL string
var port string
var defaultPort = "8080"
var host string
var envVariableOk bool

var db DBClient

//initializeEnv sets up all environment variables and prints warnings if something is missing
func initializeEnv() {

	var dbErr error
	db, dbErr = datastore.NewClient()
	if dbErr != nil {
		fmt.Println("Error instntiating Database.")
	}

	port, envVariableOk = os.LookupEnv("PORT")
	if !envVariableOk {
		fmt.Println("PORT not set.")
	}

	host, envVariableOk = os.LookupEnv("HOST_URI")
	if !envVariableOk {
		fmt.Println("HOST_URI not set.")
	}
	fmt.Println("Environment Initialized")
}

func main() {

	initializeEnv()

	router := NewRouter()

	if !(port == "") {
		fmt.Println("atmzr started and listening on port " + port)
		log.Fatal(http.ListenAndServe(":"+port, router))
	} else {
		fmt.Println("atmzr started and listening on port " + defaultPort)
		log.Fatal(http.ListenAndServe(":"+defaultPort, router))
	}
}
