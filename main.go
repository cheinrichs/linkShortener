package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var dbURL string
var dbError bool
var port string
var portError bool
var host string
var hostError bool

type Timestamp time.Time

type Response struct {
	Status string `json:"status,omitempty"`
	Data   string `json:"data,omitempty"`
}

func dbConn() (db *sql.DB) {
	db, err := sql.Open("postgres", dbURL)
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

	statisticsSQL := `INSERT INTO link_statistics (link_id)
					 VALUES ($1)`

	_, statisticsErr := db.Exec(statisticsSQL, decodedString[0])
	if statisticsErr != nil {
		panic(statisticsErr)
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func createLinkEndpoint(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		// Handle error here via logging and then return
	}

	//TODO: sanitize the data
	db := dbConn()
	defer db.Close()

	var id int
	link := r.PostFormValue("url")

	fmt.Println(link)

	sqlStatement := `INSERT INTO links (url)
					 VALUES ($1)
					 RETURNING id`

	queryErr := db.QueryRow(sqlStatement, link).Scan(&id)
	if queryErr != nil {
		panic(queryErr)
	}

	encodedString := base64.URLEncoding.EncodeToString([]byte(string(id)))

	response := Response{
		Status: "success",
		Data:   encodedString,
	}

	json.NewEncoder(w).Encode(response)
}

func linkStatisticsEndpoint(w http.ResponseWriter, r *http.Request) {

	type Response struct {
		Status string `json:"status,omitempty"`
		Data   int    `json:"data,omitempty"`
	}

	vars := mux.Vars(r)

	var decodedString, _ = base64.StdEncoding.DecodeString(vars["redirectHash"])
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

func linkTimeSeriesEndpoint(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Status string   `json:"status,omitempty"`
		Data   []string `json:"data,omitempty"`
	}

	vars := mux.Vars(r)

	var decodedString, _ = base64.StdEncoding.DecodeString(vars["redirectHash"])
	db := dbConn()
	defer db.Close()

	sqlStatement := `SELECT viewed_at FROM link_statistics WHERE link_id=$1;`

	rows, err := db.Query(sqlStatement, decodedString[0])
	defer rows.Close()

	data := make([]string, 0)

	for rows.Next() {
		var viewedAt Timestamp

		err = rows.Scan(&viewedAt)
		if err != nil {
			panic(err)
		}

		stamp := fmt.Sprintf("\"%s\"", time.Time(viewedAt).Format("2 Jan 2006 15:04:05 -0700 MST"))
		data = append(data, stamp)
	}
	if err != nil {
		panic(err)
	}

	response := Response{
		Status: "success",
		Data:   data,
	}

	json.NewEncoder(w).Encode(response)
}

func main() {

	dbURL, dbError = os.LookupEnv("DATABASE_URL")
	port, portError = os.LookupEnv("PORT")
	defaultPort := "8080"

	router := mux.NewRouter()

	router.HandleFunc("/createLink", createLinkEndpoint).Methods("POST")
	router.HandleFunc("/linkStatistics/{redirectHash}", linkStatisticsEndpoint).Methods("GET")
	router.HandleFunc("/linkTimeSeries/{redirectHash}", linkTimeSeriesEndpoint).Methods("GET")
	router.HandleFunc("/{redirectHash}", redirectEndpoint).Methods("GET")

	if !(port == "") {
		log.Fatal(http.ListenAndServe(":"+port, router))
	} else {
		log.Fatal(http.ListenAndServe(":"+defaultPort, router))
	}
}
