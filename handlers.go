package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"unicode/utf8"

	"github.com/gorilla/mux"
)

type response struct {
	Status string `json:"status,omitempty"`
	Data   string `json:"data,omitempty"`
}

var indexTemplate = template.Must(template.ParseFiles("index.tmpl"))

func indexEndpoint(w http.ResponseWriter, r *http.Request) {
	if err := indexTemplate.Execute(w, ""); err != nil {
		fmt.Println(err)
	}
}

//createLinkEndpoint /createLink makes a redirect link in the database
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

	id, insertErr := db.InsertURL(link)
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

//linkStatisticsEndpoint takes a hash and returns a count of how many times a link has been viewed
func linkStatisticsEndpoint(w http.ResponseWriter, r *http.Request) {

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

	count, clientErr := db.GetLinkViewCount(decodedString)
	if clientErr != nil {
		response := response{
			Status: "error",
			Data:   clientErr.Error(),
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

//linkStatisticsErrorEndpoint returns an error if a user hits the linkStatistics endpoint without a hash
func linkStatisticsErrorEndpoint(w http.ResponseWriter, r *http.Request) {
	response := response{
		Status: "error",
		Data:   "Please include a hash.",
	}

	json.NewEncoder(w).Encode(response)
}

//redirectEndpoint records a view statistic and redirects the user to a the requested link
//redirectHash is the id of the links table base64 encoded
func redirectEndpoint(w http.ResponseWriter, r *http.Request) {
	var requestVars = mux.Vars(r)

	var decodedString, _ = base64.StdEncoding.DecodeString(requestVars["redirectHash"])
	var linkID = decodedString[0]

	url, findErr := db.FindRedirectURLByID(linkID)

	if findErr != nil {
		http.Redirect(w, r, host, http.StatusSeeOther)
	}

	recordViewErr := db.RecordView(linkID)
	if recordViewErr != nil {
		fmt.Println(recordViewErr.Error())
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
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
