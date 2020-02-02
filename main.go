package main

import (
	"database/sql"
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
	Data   int    `json:"data,omitempty"`
}

var host, hostError = os.LookupEnv("HOST")
var port, portError = os.LookupEnv("PORT")
var user, userError = os.LookupEnv("USER")
var dbname, dbnameError = os.LookupEnv("DBNAME")

func createLinkEndpoint(w http.ResponseWriter, req *http.Request) {

	req.ParseForm()

	fmt.Println(req.FormValue("url"))
	//TODO: check to see if it's null and if not return a false
	//TODO: sanitize the data

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

	response := Response{
		Status: "Success",
		Data:   id,
	}

	json.NewEncoder(w).Encode(response)
}

func main() {

	router := mux.NewRouter()

	//TODO: implement GET redirect
	router.HandleFunc("/createlink", createLinkEndpoint).Methods("POST")

	log.Fatal(http.ListenAndServe(":12345", router))
}
