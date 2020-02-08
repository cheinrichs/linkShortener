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
	"strconv"
	"unicode/utf8"

	"github.com/cheinrichs/linkShortener/datastore"
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

type response struct {
	Status string `json:"status,omitempty"`
	Data   string `json:"data,omitempty"`
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

	dbClient, dbErr := datastore.NewClient()
	if dbErr != nil {
		fmt.Println(dbErr.Error())
		return "", dbErr
	}

	result, recordViewErr := dbClient.FindRedirectURLByID(linkID)
	if recordViewErr != nil {
		fmt.Println(recordViewErr.Error())
		return "", recordViewErr
	}
	return result, nil
}

//recordView increments the view statistics by adding a record to the link_statistics table
func recordView(linkID byte) error {

	dbClient, dbErr := datastore.NewClient()
	if dbErr != nil {
		fmt.Println(dbErr.Error())
		return dbErr
	}

	recordViewErr := dbClient.RecordView(linkID)
	if recordViewErr != nil {
		fmt.Println(recordViewErr.Error())
		return recordViewErr
	}
	return nil
}

//createLinkEndpoint
func createLinkEndpoint(w http.ResponseWriter, r *http.Request) {

	parseErr := r.ParseForm()
	if parseErr != nil {
		response := response{
			Status: "error",
			Data:   "There was a problem parsing your request.",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	link := r.PostFormValue("url")

	if link == "" {
		response := response{
			Status: "error",
			Data:   "No link provided.",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	_, urlError := url.ParseRequestURI(link)
	if urlError != nil {
		response := response{
			Status: "error",
			Data:   "Invalid URL provided.",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	id, insertErr := insertURL(link)
	if insertErr != nil {
		response := response{
			Status: "error",
			Data:   "There was a problem creating this redirect.",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	encodedString := EncodeID(id)

	response := response{
		Status: "success",
		Data:   host + encodedString,
	}

	json.NewEncoder(w).Encode(response)
	return
}

//EncodeID returns the base64 string version of the link ID
func EncodeID(id int) string {
	return base64.URLEncoding.EncodeToString([]byte(string(id)))
}

//DecodeID returns the integer linkID from a base64 encoded string
func DecodeID(id string) (int, error) {
	decoded, err := base64.StdEncoding.DecodeString(id)
	return int(decoded[0]), err
}

//insertURL actually does the db insert when creating a shortened link
func insertURL(url string) (int, error) {

	dbClient, dbErr := datastore.NewClient()
	if dbErr != nil {
		fmt.Println(dbErr.Error())
		return 0, dbErr
	}

	id, clientErr := dbClient.InsertURL(url)
	if clientErr != nil {
		fmt.Println(clientErr.Error())
		return -1, clientErr
	}
	return id, nil
}

//LinkStatisticsEndpoint takes a hash and returns a count of how many times a link has been viewed
func LinkStatisticsEndpoint(w http.ResponseWriter, r *http.Request) {

	var requestVars = mux.Vars(r)

	if requestVars["redirectHash"] == "" || utf8.RuneCountInString(requestVars["redirectHash"]) < 4 {
		response := response{
			Status: "error",
			Data:   "Please provide a valid hash.",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	var decodedString, _ = DecodeID(requestVars["redirectHash"])

	count, countError := getLinkViewCount(decodedString)
	if countError != nil {
		response := response{
			Status: "error",
			Data:   countError.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := response{
		Status: "success",
		Data:   strconv.Itoa(count),
	}

	json.NewEncoder(w).Encode(response)
}

//getLinkViewCount queries the view data for total number of times a link has been viewed
func getLinkViewCount(id int) (int, error) {
	dbClient, dbErr := datastore.NewClient()
	if dbErr != nil {
		fmt.Println(dbErr.Error())
		return 0, dbErr
	}

	count, clientErr := dbClient.GetLinkViewCount(id)
	if clientErr != nil {
		fmt.Println(clientErr.Error())
		return -1, clientErr
	}
	return count, nil
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
	router.HandleFunc("/linkStatistics/{redirectHash}", LinkStatisticsEndpoint).Methods("GET")
	router.HandleFunc("/{redirectHash}", redirectEndpoint).Methods("GET")

	if !(port == "") {
		fmt.Println("atmzr started and listening on port " + port)
		log.Fatal(http.ListenAndServe(":"+port, router))
	} else {
		fmt.Println("atmzr started and listening on port " + defaultPort)
		log.Fatal(http.ListenAndServe(":"+defaultPort, router))
	}
}
