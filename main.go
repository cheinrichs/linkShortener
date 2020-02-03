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

func redirectEndpoint(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	var decodedByte, _ = base64.StdEncoding.DecodeString(vars["redirectHash"])
	var decodedString = string(decodedByte)
	var url string

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
		host, port, user, dbname)
	db, dbErr := sql.Open("postgres", psqlInfo)
	if dbErr != nil {
		panic(dbErr)
	}
	defer db.Close()

	sqlStatement := `SELECT url FROM links WHERE id=$1;`

	row := db.QueryRow(sqlStatement, decodedString)
	err := row.Scan(&url)
	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
	case nil:
		fmt.Println(url)
	default:
		panic(err)
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func createLinkEndpoint(w http.ResponseWriter, req *http.Request) {

	req.ParseForm()

	fmt.Println(req.FormValue("url"))
	//TODO: check to see if it's null and if not return a false
	//TODO: sanitize the data
	//TODO: return the correct hash

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
		host, port, user, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var id int
	link := req.FormValue("url")

	sqlStatement := `INSERT INTO links (url) 
					 VALUES ($1)
					 RETURNING id`

	err = db.QueryRow(sqlStatement, link).Scan(&id)
	if err != nil {
		panic(err)
	}

	fmt.Println("New record ID is:", id)

	encodedString := base64.URLEncoding.EncodeToString([]byte(string(id)))
	fmt.Printf("Encoded: %s\n", encodedString)

	raw, err := base64.URLEncoding.DecodeString(encodedString)
	if err != nil {
		panic(err)
	}
	fmt.Println("Decoded:", raw)

	response := Response{
		Status: "Success",
		Data:   encodedString,
	}

	json.NewEncoder(w).Encode(response)
}

func main() {

	router := mux.NewRouter()

	//TODO: implement GET redirect
	//TODO: implement GET stats

	router.HandleFunc("/{redirectHash}", redirectEndpoint).Methods("GET")
	router.HandleFunc("/createlink", createLinkEndpoint).Methods("POST")

	log.Fatal(http.ListenAndServe(":12345", router))
}
