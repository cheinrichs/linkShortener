package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var dbURL string
var port string
var host string

var envVariableOk bool
var defaultPort = "8080"

var db *sql.DB
var dbErr error

//initializeEnv sets up all environment variables and prints warnings if something is missing
func initializeEnv() {
	dbURL, envVariableOk = os.LookupEnv("DATABASE_URL")
	if !envVariableOk {
		fmt.Println("DATABASE_URL not set.")
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

	router := mux.NewRouter()

	router.HandleFunc("/createLink", createLinkEndpoint).Methods("POST")
	router.HandleFunc("/linkStatistics/{redirectHash}", linkStatisticsEndpoint).Methods("GET")
	router.HandleFunc("/{redirectHash}", redirectEndpoint).Methods("GET")

	if !(port == "") {
		fmt.Println("atmzr started and listening on port " + port)
		log.Fatal(http.ListenAndServe(":"+port, router))
	} else {
		fmt.Println("atmzr started and listening on port " + defaultPort)
		log.Fatal(http.ListenAndServe(":"+defaultPort, router))
	}
}
