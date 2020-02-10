package main

import (
	"bytes"
	"database/sql"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func Router() *mux.Router {
	return NewRouter()
}

type Mockdb struct {
	db *sql.DB
}

//FindRedirectURLByID mocked for testing
func (m Mockdb) FindRedirectURLByID(linkID byte) (string, error) {

	//Used in TestRedirectEndpointNonexistantRedirect
	if linkID == 65 {
		return "", errors.New("error: redirect not found")
	}
	return "http://www.google.com", nil
}

//RecordView mocked for testing
func (m Mockdb) RecordView(linkID byte) error {
	return nil
}

//InsertURL mocked for testing
func (m Mockdb) InsertURL(link string) (int, error) {
	return 1, nil
}

//GetLinkViewCount mocked for testing
func (m Mockdb) GetLinkViewCount(id int) (int, error) {

	//called by TestLinkStatisticsEndpointLinkDoesNotExist
	if id == 207 {
		return 0, nil
	}

	return 1, nil
}

func init() {
	db = Mockdb{db: nil}
}

//CreateLinkEndpoint
//Create a valid link
func TestCreateLinkValidURL(t *testing.T) {

	host = "http://localhost:8080/"
	u := "https://localhost:8080/createLink"
	data := url.Values{}
	data.Set("url", "http://www.google.com")

	byteString := bytes.NewBufferString(data.Encode())

	request, _ := http.NewRequest("POST", u, byteString)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := httptest.NewRecorder()

	Router().ServeHTTP(response, request)

	b, _ := ioutil.ReadAll(response.Body)
	actual := string(b)
	expected := "{\"status\":\"success\",\"data\":\"http://localhost:8080/AQ==\"}\n"

	assert.Equal(t, expected, actual, "Redirecting to incorrect url.")
}

//Create a link with an invalid URL
func TestCreateLinkInvalidURL(t *testing.T) {

	host = "http://localhost:8080/"
	u := "https://localhost:8080/createLink"
	data := url.Values{}
	data.Set("url", "http//www.google.com")

	byteString := bytes.NewBufferString(data.Encode())

	request, _ := http.NewRequest("POST", u, byteString)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := httptest.NewRecorder()

	Router().ServeHTTP(response, request)

	b, _ := ioutil.ReadAll(response.Body)
	actual := string(b)
	expected := "{\"status\":\"error\",\"data\":\"Invalid URL provided.\"}\n"

	assert.Equal(t, expected, actual, "Redirecting to incorrect url.")
}

//CreateLink that has no url
func TestCreateLinkMissingURL(t *testing.T) {

	host = "http://localhost:8080/"
	u := "https://localhost:8080/createLink"
	data := url.Values{}
	data.Set("url", "")

	byteString := bytes.NewBufferString(data.Encode())

	request, _ := http.NewRequest("POST", u, byteString)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := httptest.NewRecorder()

	Router().ServeHTTP(response, request)

	b, _ := ioutil.ReadAll(response.Body)
	actual := string(b)
	expected := "{\"status\":\"error\",\"data\":\"No link provided.\"}\n"

	assert.Equal(t, expected, actual, "Redirecting to incorrect url.")
}

//RedirectEndpoint
func TestRedirectEndpointNonexistantRedirect(t *testing.T) {

	request, _ := http.NewRequest("GET", "/QQ==", nil)
	response := httptest.NewRecorder()

	Router().ServeHTTP(response, request)
	expected := "/"
	actual := response.Header()["Location"][0]

	assert.Equal(t, expected, actual, "Redirecting to incorrect url.")
}

//Should redirect to Index
func TestRedirectEndpointValidRedirect(t *testing.T) {

	request, _ := http.NewRequest("GET", "/SQ==", nil)
	response := httptest.NewRecorder()

	Router().ServeHTTP(response, request)
	actual := response.Header()["Location"][0]

	expected := "http://www.google.com"
	assert.Equal(t, expected, actual, "Redirecting to incorrect url.")
}

//IndexEndpoint
func TestIndexEndpoint(t *testing.T) {

	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	Router().ServeHTTP(response, request)

	b, _ := ioutil.ReadAll(response.Body)
	actual := string(b)
	expected := "<!DOCTYPE html>\n<html>\n\t<head>\n\t\t<meta charset=\"utf-8\">\n\t\t<title>atmzr</title>\n\t\t<style>\n\t\t\tbody {\n\t\t\t\tfont-family: sans-serif;\n\t\t\t}\n\t\t</style>\n\t</head>\n\t<body>\n\t\t<h1>atmzr</h1>\n        <p>Welcome to atmzr!</p>\n        <p>Please visit <a href=\"https://github.com/cheinrichs/linkShortener\">https://github.com/cheinrichs/linkShortener</a> for more information.</p>\n\t</body>\n</html>"
	assert.Equal(t, expected, actual, "Index does not return the correct HTML.")
}

//LinkStatisticsEndpoint
//Test link exists and returns correct data
func TestLinkStatisticsEndpointLinkExistsWithData(t *testing.T) {

	request, _ := http.NewRequest("GET", "/linkStatistics/SQ==", nil)
	response := httptest.NewRecorder()

	Router().ServeHTTP(response, request)

	b, _ := ioutil.ReadAll(response.Body)

	actual := string(b)
	expected := "{\"status\":\"success\",\"data\":\"1\"}\n"

	assert.Equal(t, expected, actual, "Does not find the correct data.")
}

//Test link does not exist
func TestLinkStatisticsEndpointLinkDoesNotExist(t *testing.T) {
	request, _ := http.NewRequest("GET", "/linkStatistics/z6k=", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)

	b, _ := ioutil.ReadAll(response.Body)

	actual := string(b)
	expected := "{\"status\":\"success\",\"data\":\"0\"}\n"

	assert.Equal(t, expected, actual, "Fails to find 0 views for a link that does not exist.")
}

//Test empty redirect hash
func TestLinkStatisticsEndpointEmptyRedirectHash(t *testing.T) {
	request, _ := http.NewRequest("GET", "/linkStatistics", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)

	b, _ := ioutil.ReadAll(response.Body)

	actual := string(b)
	expected := "{\"status\":\"error\",\"data\":\"Please include a hash.\"}\n"

	assert.Equal(t, expected, actual, "Fails to alert the user to include a hash.")
}

//Test redirect hash that's too small
func TestLinkStatisticsEndpointRedirectHashTooSmall(t *testing.T) {
	request, _ := http.NewRequest("GET", "/linkStatistics/SQ=", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)

	b, _ := ioutil.ReadAll(response.Body)
	actual := string(b)
	expected := "{\"status\":\"error\",\"data\":\"Please provide a valid hash.\"}\n"

	assert.Equal(t, expected, actual, "Fails to alert the user to include a hash.")
}

//TestDecodeID makes sure we decode numbers correctly
func TestDecodeID(t *testing.T) {
	var input = "SQ=="
	var expected = 73
	actual, err := DecodeID(input)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	assert.Equal(t, expected, actual, "Decoding working incorrectly")
}

//TestEncodeID makes sure we encode numbers correctly
func TestEncodeID(t *testing.T) {
	var input = 73
	var expected = "SQ=="
	actual := EncodeID(input)

	assert.Equal(t, expected, actual, "Encoding working incorrectly")
}
