package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	count, clientErr := db.GetLinkViewCount(id)
	if clientErr != nil {
		fmt.Println(clientErr.Error())
		return -1, clientErr
	}
	return count, nil
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

	result, recordViewErr := db.FindRedirectURLByID(linkID)
	if recordViewErr != nil {
		fmt.Println(recordViewErr.Error())
		return "", recordViewErr
	}
	return result, nil
}

//recordView increments the view statistics by adding a record to the link_statistics table
func recordView(linkID byte) error {

	recordViewErr := db.RecordView(linkID)
	if recordViewErr != nil {
		fmt.Println(recordViewErr.Error())
		return recordViewErr
	}
	return nil
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
