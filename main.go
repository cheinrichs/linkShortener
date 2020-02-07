package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var dbURL string
var port string
var host string

var envVariableOk bool
var defaultPort = "8080"

type Response struct {
	Status string `json:"status,omitempty"`
	Data   string `json:"data,omitempty"`
}

type errorResponse struct {
	Status string `json:"status,omitempty"`
	Error  error  `json:"data,omitempty"`
}

//dbConn connects to the database
func dbConn() (db *sql.DB) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}
	return db
}

//redirectEndpoint records a view statistic and redirects the user to a the requested link
//redirectHash is the id of the links table base64 encoded
func redirectEndpoint(w http.ResponseWriter, r *http.Request) {
	var requestVars = mux.Vars(r)

	var decodedString, _ = base64.StdEncoding.DecodeString(requestVars["redirectHash"])
	var linkID = decodedString[0]

	url, findErr := findRedirectURLByID(linkID)

	if findErr != nil {
		http.Redirect(w, r, host, http.StatusSeeOther)
	}

	recordViewErr := recordView(linkID)
	if recordViewErr != nil {
		panic(recordViewErr)
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

//findRedirectURLByID returns the record in the database with the given ID
func findRedirectURLByID(linkID byte) (string, error) {
	var result string

	db := dbConn()
	defer db.Close()

	sqlStatement := `SELECT url FROM links WHERE id=$1;`

	row := db.QueryRow(sqlStatement, linkID)
	err := row.Scan(&result)
	switch err {
	case sql.ErrNoRows:
		fmt.Println("no rows")
		return "", nil
	case nil:
		fmt.Println(result)
		return result, nil
	default:
		fmt.Println("something else")
		return "", err
	}
}

//recordView increments the view statistics by adding a record to the link_statistics table
func recordView(linkID byte) error {

	db := dbConn()
	defer db.Close()

	statisticsSQL := `INSERT INTO link_statistics (link_id)
					 VALUES ($1)`

	_, statisticsErr := db.Exec(statisticsSQL, linkID)

	return statisticsErr
}

//createLinkEndpoint
func createLinkEndpoint(w http.ResponseWriter, r *http.Request) {

	parseErr := r.ParseForm()
	if parseErr != nil {
		response := Response{
			Status: "error",
			Data:   "There was a problem parsing your request.",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	link := r.PostFormValue("url")

	_, urlError := url.ParseRequestURI(link)
	if urlError != nil {
		response := Response{
			Status: "error",
			Data:   "Invalid URL provided.",
		}
		json.NewEncoder(w).Encode(response)
	}

	id, insertErr := insertURL(link)
	if insertErr != nil {
		response := Response{
			Status: "error",
			Data:   "There was a problem creating this redirect.",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	encodedString := encodeID(id)

	response := Response{
		Status: "success",
		Data:   host + encodedString,
	}

	json.NewEncoder(w).Encode(response)
	return
}

//encodeID returns the base64 string version of the link ID
func encodeID(id int) string {
	return base64.URLEncoding.EncodeToString([]byte(string(id)))
}

func insertURL(link string) (int, error) {
	var id int

	db := dbConn()
	defer db.Close()

	sqlStatement := `INSERT INTO links (url)
					 VALUES ($1)
					 RETURNING id`

	queryErr := db.QueryRow(sqlStatement, link).Scan(&id)

	return id, queryErr
}

//linkStatisticsEndpoint returns a count of how many times a link has been viewed
func linkStatisticsEndpoint(w http.ResponseWriter, r *http.Request) {

	type Response struct {
		Status string `json:"status,omitempty"`
		Data   int    `json:"data,omitempty"`
	}

	var requestVars = mux.Vars(r)

	var decodedString, _ = base64.StdEncoding.DecodeString(requestVars["redirectHash"])

	db := dbConn()
	defer db.Close()

	sqlStatement := `SELECT COUNT(*) FROM link_statistics WHERE link_id=$1;`

	var count int

	row := db.QueryRow(sqlStatement, decodedString[0])
	err := row.Scan(&count)
	switch err {
	case sql.ErrNoRows:
		count = 0
	case nil:
		fmt.Println(count)
	default:
		panic(err)
	}

	response := Response{
		Status: "success",
		Data:   count,
	}

	json.NewEncoder(w).Encode(response)
}

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
