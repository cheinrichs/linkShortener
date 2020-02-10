package main

import (
	"database/sql"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
	return "", nil
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

//RedirectEndpoint

//IndexEndpoint
func TestIndexEndpoint(t *testing.T) {

	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	Router().ServeHTTP(response, request)

	b, _ := ioutil.ReadAll(response.Body)
	expected := "<!DOCTYPE html>\n<html>\n\t<head>\n\t\t<meta charset=\"utf-8\">\n\t\t<title>atmzr</title>\n\t\t<style>\n\t\t\tbody {\n\t\t\t\tfont-family: sans-serif;\n\t\t\t}\n\t\t</style>\n\t</head>\n\t<body>\n\t\t<h1>atmzr</h1>\n        <p>Welcome to atmzr!</p>\n        <p>Please visit <a href=\"https://github.com/cheinrichs/linkShortener\">https://github.com/cheinrichs/linkShortener</a> for more information.</p>\n\t</body>\n</html>"
	assert.Equal(t, expected, string(b), "Existing link returns correct data.")
}

//LinkStatisticsEndpoint
//Test link exists and returns correct data
func TestLinkStatisticsEndpointLinkExistsWithData(t *testing.T) {

	request, _ := http.NewRequest("GET", "/linkStatistics/SQ==", nil)
	response := httptest.NewRecorder()

	Router().ServeHTTP(response, request)

	b, _ := ioutil.ReadAll(response.Body)

	assert.Equal(t, "{\"status\":\"success\",\"data\":\"1\"}\n", string(b), "Existing link returns correct data.")
}

//Test link does not exist
func TestLinkStatisticsEndpointLinkDoesNotExist(t *testing.T) {
	request, _ := http.NewRequest("GET", "/linkStatistics/z6k=", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)

	b, _ := ioutil.ReadAll(response.Body)

	assert.Equal(t, "{\"status\":\"success\",\"data\":\"0\"}\n", string(b), "Existing link returns correct data.")
}

//Test empty redirect hash
func TestLinkStatisticsEndpointEmptyRedirectHash(t *testing.T) {
	request, _ := http.NewRequest("GET", "/linkStatistics", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)

	b, _ := ioutil.ReadAll(response.Body)

	assert.Equal(t, "{\"status\":\"error\",\"data\":\"Please include a hash.\"}\n", string(b), "Existing link returns correct data.")
}

//Test redirect hash that's too small
func TestLinkStatisticsEndpointRedirectHashTooSmall(t *testing.T) {
	request, _ := http.NewRequest("GET", "/linkStatistics/SQ=", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)

	b, _ := ioutil.ReadAll(response.Body)
	assert.Equal(t, "{\"status\":\"error\",\"data\":\"Please provide a valid hash.\"}\n", string(b), "Existing link returns correct data.")
}

//TestDecodeID makes sure we decode numbers correctly
func TestDecodeID(t *testing.T) {
	var input = "SQ=="
	var expected = 73
	result, err := DecodeID(input)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	assert.Equal(t, expected, result, "Expecting `73`")
}

//TestEncodeID makes sure we encode numbers correctly
func TestEncodeID(t *testing.T) {
	var input = 73
	var expected = "SQ=="
	result := EncodeID(input)

	assert.Equal(t, expected, result, "Expecting `SQ==`")
}
