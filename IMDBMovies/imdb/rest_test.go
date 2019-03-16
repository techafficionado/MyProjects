/******************************************************************************
 * \file        rest_test.go
 *
 * \brief       GO File that has Database Model and Access Object defined
 *
 * \author      Reshma Syeda
 *
 * ****************************************************************************/
// Enhancements: Add stretchr/assert library
// Enhancements: Make separate package main_test
package main

import(
		"testing"
		"net/http"
		"net/http/httptest"
		"github.com/gorilla/mux"
		//"fmt"
		"encoding/json"
		"io/ioutil"
		"io"
		"os"
		"bytes"
		"mime/multipart"
		log "github.com/sirupsen/logrus"
		"strings"
)

// ErrorJSON struct for error responses
type ErrorJSON struct{
	Code string `json:"code"`
	ErrorMsg string `json:"error"`
}

// Database Access Object for Testing
var dao_test = MoviesDAO{}

/******************************************************************************************
 *
 * Build HTTP Router
 *
*******************************************************************************************/
func Router() *mux.Router{

	dao_test.Server = "localhost"
    dao_test.Database =  "MoviesDB"
    dao_test.Connect()


	// Disable logging
	discard := io.MultiWriter()
	log.SetOutput(discard)

	router := mux.NewRouter()
    router.HandleFunc("/imdb/version", GetVersion).Methods("GET") // get version
    router.HandleFunc("/imdb/uploadmovies", PostCSV).Methods("POST") // post movie uploads
    router.HandleFunc("/imdb/movies", GetMovies).Methods("GET") // get movies
    router.HandleFunc("/imdb/endpoints", GetEndpoints).Methods("GET") // get Rest Endpoint

	return router

}
/******************************************************************************************
 *
 * Setup Upload Request
 *
*******************************************************************************************/
func SetUploadRequest(uri string, path string, paramName string, setHeader bool)(*http.Request, error){
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, fi.Name())
	if err != nil {
		return nil, err
	}
	part.Write(fileContents)

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request,_ := http.NewRequest("POST","/imdb/uploadmovies", body)

	if setHeader == true{
		request.Header.Add("Content-Type", writer.FormDataContentType())
	}
	return request,err
}

/******************************************************************************************
 *
 * Test for ERR_CONTENT_TYPE_INVALID
 *
*******************************************************************************************/
func TestPostCSVContentType(t *testing.T) {

	path := "./test/IMDB-Movie-Data_Assignment.csv"
	paramName := "file"

	req,_ := SetUploadRequest("/imdb/uploadmovies",path,paramName, false) 


	resp := httptest.NewRecorder()
	Router().ServeHTTP(resp, req)

	var errjson = new(ErrorJSON)
	err1 := json.NewDecoder(resp.Body).Decode(errjson)

	if (err1 != nil ||
		resp.Code != 415 ||
		errjson.ErrorMsg != "Please upload file as multipart/form-data with \"file\" as key"){
		t.Errorf("TestPostCSV Failed")
	}

}

/******************************************************************************************
 *
 * Test for ERR_FILE_TOO_BIG
 *
*******************************************************************************************/
func TestPostCSVFileSize(t *testing.T) {

	path := "./test/largefile.csv"
	paramName := "file"

	req,_ := SetUploadRequest("/imdb/uploadmovies",path,paramName, true)

	resp := httptest.NewRecorder()
	Router().ServeHTTP(resp, req)

	//fmt.Println(resp.Body)

	var errjson = new(ErrorJSON)
	err1 := json.NewDecoder(resp.Body).Decode(errjson)

	if (err1 != nil ||
		resp.Code != 400 ||
		!strings.Contains(errjson.ErrorMsg, "File is too large. Maximum upload size")){
		t.Errorf("TestPostCSVFileSize Failed")
	}

}

/******************************************************************************************
 *
 * Test for successful record creation
 *
*******************************************************************************************/
func TestPostCSVCreate(t *testing.T) {
	dao_test.Clean()
	path := "./test/passlist.csv"
	paramName := "file"
	req,_ := SetUploadRequest("/imdb/uploadmovies",path,paramName,true)

	resp := httptest.NewRecorder()
	Router().ServeHTTP(resp, req)

	var jres = new(UploadResults)
	err := json.NewDecoder(resp.Body).Decode(jres)

	if (err != nil || resp.Code != 200){
		t.Errorf("TestPostCSVCreate Failed")
	}

	if jres.RecordsCreated != 5{
		t.Errorf("TestPostCSVCreate Failed")
	}
}

/******************************************************************************************
 *
 * Test for Errored record in case of duplicate upload entries
 *
*******************************************************************************************/
func TestPostCSVDuplicate(t *testing.T) {
	path := "./test/passlist.csv"
	paramName := "file"
	req,_ := SetUploadRequest("/imdb/uploadmovies",path,paramName,true)

	resp := httptest.NewRecorder()
	Router().ServeHTTP(resp, req)

	var jres = new(UploadResults)
	err := json.NewDecoder(resp.Body).Decode(jres)


	if (err != nil || resp.Code != 200){
		t.Errorf("TestPostCSVCreate Failed")
	}

	if jres.RecordsErrored != 5{
		t.Errorf("TestPostCSVDuplicate Failed")
	}
}

/******************************************************************************************
 *
 * Test for correct csv file format with 12 columns
 *
*******************************************************************************************/
func TestPostCSVFormat(t *testing.T) {
	path := "./test/faillist.csv"
	paramName := "file"
	req,_ := SetUploadRequest("/imdb/uploadmovies",path,paramName,true)

	resp := httptest.NewRecorder()
	Router().ServeHTTP(resp, req)

	var errjson = new(ErrorJSON)
	err := json.NewDecoder(resp.Body).Decode(errjson)

	if (err != nil || resp.Code != 400 ||
		errjson.ErrorMsg != "Invalid File Format"){
		t.Errorf("TestPostCSVFormat Failed")
	}
}

/******************************************************************************************
 *
 * Test for ERR_YEAR_INVALID
 *
*******************************************************************************************/
func TestGetByInvalidYear(t *testing.T) {

	req,err := http.NewRequest("GET","/imdb/movies?year=20155",nil)
	resp := httptest.NewRecorder()
	Router().ServeHTTP(resp, req)
	var errjson = new(ErrorJSON)
	err = json.NewDecoder(resp.Body).Decode(errjson)

	if (err != nil ||
		resp.Code != 400 ||
		errjson.ErrorMsg != "Please provide a valid year"){
		t.Errorf("TestGetByInvalidYear Failed")
	}
}


/******************************************************************************************
 *
 * Test for Invalid Genre ERR_GENRE_INVALID
 *
*******************************************************************************************/
func TestGetByInvalidGenre(t *testing.T) {
	req,err := http.NewRequest("GET","/imdb/movies?genre=",nil)
	resp := httptest.NewRecorder()
	Router().ServeHTTP(resp, req)

	var errjson = new(ErrorJSON)
	err = json.NewDecoder(resp.Body).Decode(errjson)

	if (err != nil ||
		resp.Code != 400 ||
		errjson.ErrorMsg != "Please provide a valid genre"){
		t.Errorf("TestGetByInvalidGenre Failed")
	}
}

/******************************************************************************************
 *
 * Test for ERR_YEAR_AND_RANGE
 *
*******************************************************************************************/
func TestGetByYearAndYearRange(t *testing.T) {
	req,err := http.NewRequest("GET","/imdb/movies?year=2016&year_from=2015",nil)
	resp := httptest.NewRecorder()
	Router().ServeHTTP(resp, req)

	var errjson = new(ErrorJSON)
	err = json.NewDecoder(resp.Body).Decode(errjson)

	if (err != nil ||
		resp.Code != 400 ||
		errjson.ErrorMsg != "Please provide either the year or a range but not both"){
		t.Errorf("TestGetByYearAndYearRange Failed")
	}
}

/******************************************************************************************
 *
 * Test for Valid Year Range ERR_YEAR_RANGE_INVALID
 *
*******************************************************************************************/
func TestGetByChronoOrder(t *testing.T) {
	req,err := http.NewRequest("GET","/imdb/movies?year_from=2014&year_to=2013",nil)
	resp := httptest.NewRecorder()
	Router().ServeHTTP(resp, req)

	var errjson = new(ErrorJSON)
	err = json.NewDecoder(resp.Body).Decode(errjson)

	if (err != nil ||
		resp.Code != 400 ||
		errjson.ErrorMsg != "Please provide a valid year_from and year_to in chronological order"){
		t.Errorf("TestGetByChronoOrder Failed")
	}
}


