package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

//Response is an object returned to the caller that contains the requested data, error information, and a success status
type Response struct {
	Status string `json:"status,omitempty"`
	Data   string `json:"data,omitempty"`
}

var host, hostError = os.LookupEnv("HOST")
var port, portError = os.LookupEnv("PORT")
var user, userError = os.LookupEnv("USER")
var dbname, dbnameError = os.LookupEnv("DBNAME")

func dbConn() (db *sql.DB) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
		host, port, user, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	return db
}

func redirectEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var decodedString, _ = base64.StdEncoding.DecodeString(vars["redirectHash"])

	var url string

	db := dbConn()
	defer db.Close()

	sqlStatement := `SELECT url FROM links WHERE id=$1;`

	row := db.QueryRow(sqlStatement, decodedString[0])
	err := row.Scan(&url)
	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
	case nil:
		fmt.Println(url)
	default:
		panic(err)
	}

	statisticsSQL := `INSERT INTO link_statistics (link_id, viewtime)
					 VALUES ($1, current_timestamp)`

	_, statisticsErr := db.Exec(statisticsSQL, decodedString[0])
	if statisticsErr != nil {
		panic(statisticsErr)
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func createLinkEndpoint(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	//TODO: sanitize the data
	db := dbConn()
	defer db.Close()

	var id int
	link := r.FormValue("url")

	sqlStatement := `INSERT INTO links (url)
					 VALUES ($1)
					 RETURNING id`

	err := db.QueryRow(sqlStatement, link).Scan(&id)
	if err != nil {
		panic(err)
	}

	encodedString := base64.URLEncoding.EncodeToString([]byte(string(id)))

	response := Response{
		Status: "Success",
		Data:   encodedString,
	}

	json.NewEncoder(w).Encode(response)
}

func linkStatisticsEndpoint(w http.ResponseWriter, r *http.Request) {

	db := dbConn()
	defer db.Close()

	response := Response{
		Status: "Success",
		Data:   "",
	}

	json.NewEncoder(w).Encode(response)
}

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/createLink", createLinkEndpoint).Methods("POST")
	router.HandleFunc("/linkStatistics/{redirectHash}", linkStatisticsEndpoint).Methods("GET")
	router.HandleFunc("/{redirectHash}", redirectEndpoint).Methods("GET")

	log.Fatal(http.ListenAndServe(":12345", router))
}
