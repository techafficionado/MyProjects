/******************************************************************************
 * \file        rest.go
 *
 * \brief       GO File that has REST functions defined
 *
 * \author      Reshma Syeda
 *
 * ****************************************************************************/


package main


import (
	"encoding/json"
    log "github.com/sirupsen/logrus"
    "net/http"
	"strconv"
	"io"
	"fmt"
	"strings"
	"errors"
	"encoding/csv"
    "gopkg.in/mgo.v2/bson"
)

// Movie Struct for Movie Record in CSV
type Movie struct {
	ID      bson.ObjectId `bson:"_id,omitempty"`
	Rank int `json:"rank"`
	Title string `json:"title"`
	Genre []string `json:"genre"`
	Description string `json:"description"`
	Director string `json:"director"`
	Actors string `json:"actors"`
	Year int `json:"year"`
	RuntimeMin int `json:"runtime_min"`
	Rating float64 `json:"rating"`
	Votes int `json:"votes"`
	RevenueMil float64 `json:"revenue_mil"`
	Metascore int `json:"metascore"`
}

// MovieGet Struct for Get API
type MovieGet struct {
	Title string `json:"title"`
	Genre []string `json:"genre"`
	Description string `json:"description"`
	Year int `json:"year"`
	RuntimeMin int `json:"runtime_min"`
	Rating float64 `json:"rating"`
}

// UploadResults Struct for POST Response
type UploadResults struct{
	RecordsRead int `json:"RecordsRead"`
	RecordsCreated int `json:"RecordsCreated"`
	RecordsErrored int `json:"RecordsErrored"`
}

type ErrorCode int

// Custom Error Codes
const (
    ERR_FILE_INVALID				ErrorCode = 0
    ERR_FILE_INVALID_FORMAT			ErrorCode = 1
    ERR_FILE_TOO_BIG				ErrorCode = 2
    ERR_YEAR_AND_RANGE				ErrorCode = 3
    ERR_YEAR_RANGE_INVALID			ErrorCode = 4
    ERR_YEAR_INVALID				ErrorCode = 5
	ERR_GENRE_INVALID				ErrorCode = 6
    ERR_INTERNAL_SERVER				ErrorCode = 7
    ERR_NO_CONTENT					ErrorCode = 8
	ERR_CONTENT_TYPE_INVALID		ErrorCode = 9
)

// Maximum Upload Size File Settings
var maxUploadSize = conf.Settings.FileSizeKB * 1024
var defaultMaxUploadSize = Max(2048*1024,conf.Settings.FileSizeKB * 1024)

// Max function for int64 datatypes
func Max(x, y int64) int64 {
    if x < y {
        return y
    }
    return x
}

/******************************************************************************************
 * Return Error Message given the ErrorCode
******************************************************************************************/
func ErrorMsg(ec ErrorCode) string{
    msg := ""
    switch ec {
		case ERR_FILE_INVALID:
			msg = "Invalid File"
		case ERR_FILE_INVALID_FORMAT:
			msg = "Invalid File Format"
		case ERR_FILE_TOO_BIG:
			msg = fmt.Sprintf("File is too large. Maximum upload size is %d Bytes", Max(2048*1024,conf.Settings.FileSizeKB * 1024))
		case ERR_YEAR_AND_RANGE:
			msg = "Please provide either the year or a range but not both"
		case ERR_YEAR_RANGE_INVALID:
			msg = "Please provide a valid year_from and year_to in chronological order"
		case ERR_YEAR_INVALID:
			msg = "Please provide a valid year"
		case ERR_GENRE_INVALID:
			msg = "Please provide a valid genre"
        case ERR_INTERNAL_SERVER:
            msg = "Internal Server Error"
		case ERR_CONTENT_TYPE_INVALID:
			msg = "Please upload file as multipart/form-data with \"file\" as key"
        default:
            msg = "Unknown Error Occured"
    }

    return msg
}



/******************************************************************************************
 * Return HTTP Status code given the ErrorCode
******************************************************************************************/
func HTTPCode(ec ErrorCode) int{
    code := 200
    switch ec {
        case ERR_FILE_INVALID,
             ERR_FILE_INVALID_FORMAT,
             ERR_FILE_TOO_BIG,
			 ERR_YEAR_AND_RANGE,
			 ERR_YEAR_RANGE_INVALID,
			 ERR_YEAR_INVALID,
			 ERR_GENRE_INVALID:
            code = 400
        case ERR_INTERNAL_SERVER:
            code = 500
        case ERR_NO_CONTENT:
            code = 204
		case ERR_CONTENT_TYPE_INVALID:
			code = 415
    }
    return code
}

/******************************************************************************************
 * Send JSON Response
******************************************************************************************/
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    response, _ := json.Marshal(payload)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(response)
}


/******************************************************************************************
 * Send JSON Response with Error code and message given an ErrorCode
******************************************************************************************/
func respondWithErrorCode(w http.ResponseWriter, errc ErrorCode) {
    code := HTTPCode(errc)
    msg := ErrorMsg(errc)
    respondWithJSON(w, code, map[string]string{"error":msg,"code":strconv.Itoa(code)})
}

/******************************************************************************************
 *
 * Get Version
 *
******************************************************************************************/
func GetVersion(w http.ResponseWriter, r *http.Request) {

    log.WithFields(log.Fields{"EndPoint":"GetVersion"}).Info()

	w.Header().Set("Content-Type", "application/json")

	version := map[string]interface{}{
        "name":    "IMDB Movie REST Service",
        "version": Version(),
    }

	response, err := json.MarshalIndent(version, "", "  ")
    if err != nil {
        respondWithErrorCode(w, ERR_INTERNAL_SERVER)
        return
    }
	w.Write(response)
}

/******************************************************************************************
 *
 * Get Endpoints
 *
******************************************************************************************/
func GetEndpoints(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/text")
    http.ServeFile(w, r, "data/swagger.yaml")
}

/******************************************************************************************
 *
 * Post CSV movie data file
 *
******************************************************************************************/
func PostCSV(w http.ResponseWriter, r *http.Request) {

    log.WithFields(log.Fields{"EndPoint":"PostCSV"}).Info()

	contentType := r.Header.Get("Content-type")

	log.WithFields(log.Fields{"contentType":contentType}).Info()

	if !(strings.Contains(contentType, "multipart/form-data")){
		log.Info("Content is NOT multipart/form-data")
		respondWithErrorCode(w, ERR_CONTENT_TYPE_INVALID)
		return
	}

	// Validate File size, return FILE_TOO_BIG
	maxUploadSize := Max(2048*1024,conf.Settings.FileSizeKB * 1024)
	log.WithFields(log.Fields{"maxUploadSize":maxUploadSize}).Info()
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
    if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		log.WithFields(log.Fields{"err":err}).Info()
		respondWithErrorCode(w, ERR_FILE_TOO_BIG)
        return
    }

	file, _ , err := r.FormFile("file")

	if err != nil {
		respondWithErrorCode(w, ERR_FILE_INVALID)
        return
    }

    defer file.Close()

	reader := csv.NewReader(file)
	if (err!= nil){
		log.WithFields(log.Fields{"error":err}).Info()
	}

	reader.FieldsPerRecord = 12
	reader.TrimLeadingSpace = true
	header := true

	insRecords := 0
	errRecords := 0
	totRecords := 0
	i := 0
	for {
        line, error := reader.Read()
		i += 1
        if error == io.EOF {
            break
        } else if error != nil {
            log.WithFields(log.Fields{"Invalid File Content in line. Error":error}).Info()
			// if we encounter this error in the first line - consider it as invalid file
			if (header == true){
				respondWithErrorCode(w, ERR_FILE_INVALID_FORMAT)
				return
			}
			// else just skip the line and move to next
			continue
        }

		if line == nil || header == true {
			log.WithFields(log.Fields{"Line Number":i,"Ignoring line":line}).Info()
			header = false
			continue
		}

		// ignore line if there is no rank
		if (len(line[0]) == 0){
			continue
		}

		// if title and/or year are missing - skip the record
		if (len(line[1]) == 0 || len(line[6]) == 0) {
			log.WithFields(log.Fields{"Title and/or year are missing":line}).Info()
			continue
		}

		// increment total valid records
		totRecords += 1

		movie, err := ValidateMovie(line)

		if (err!= nil){
			log.WithFields(log.Fields{"Movie Record Validation Failed for Line":line}).Info()
			errRecords += 1
			continue
		}

		// insert to db
		err = dao.Insert(*movie)
		if err != nil {
			log.WithFields(log.Fields{"Insert Error":err}).Info()
			errRecords += 1
		}else{ // insert failed
			insRecords += 1
		}
	}

	log.WithFields(log.Fields{"Total Records Created":insRecords}).Info()

	w.Header().Set("Content-Type", "application/json")

	var uploadresults = new(UploadResults)
	uploadresults.RecordsRead = totRecords
	uploadresults.RecordsCreated = insRecords
	uploadresults.RecordsErrored = errRecords

	json.NewEncoder(w).Encode(uploadresults)
}

/******************************************************************************************
 *
 * Validate Movie Record in CSV
 *
******************************************************************************************/
func ValidateMovie(line []string) (*Movie, error){

	var movie = new(Movie)

	// if rank is not present we would never reach here
	rank, err := strconv.Atoi(line[0])

	if (err!= nil){
		log.WithFields(log.Fields{"Rank conversion failed":line[0], "err":err}).Info()
		return movie,err
	}

	// split genre list to array
	genrelist := strings.Split(strings.ToLower(line[2]), ",")

	// convert year from string to int
	year, err := strconv.Atoi(line[6])

	if (err!= nil){
		log.WithFields(log.Fields{"Year conversion failed":line[6]}).Info()
		return movie,err
	}

	// convert runtime from string to int
	runtime, err := strconv.Atoi(line[7])

	if (err!= nil){
		log.WithFields(log.Fields{"RuntimeMin conversion failed":line[0]}).Info()
		return movie,err
	}

	// convert rating from string to float64
	rating, err := strconv.ParseFloat(line[8], 64)

	if (err!= nil){
		log.WithFields(log.Fields{"Rating conversion failed":line[8]}).Info()
		return movie,err
	}

	// convert votes from string to int
	votes, err := strconv.Atoi(line[9])

	if (err!= nil){
		log.WithFields(log.Fields{"Votes conversion failed":line[9]}).Info()
		return movie,err
	}

	// revenue field could be empty, needs special handling
	var revenue float64
	if (len(line[10]) == 0){
		revenue = 0.0
	}else{
		var err1 error
		// convert revenue from string to float64
		revenue, err1 = strconv.ParseFloat(line[10], 64)

		if (err1!= nil){
			log.WithFields(log.Fields{"Revenue conversion failed":line[10]}).Info()
			return movie,err1
		}
	}

	// metascore field could be empty, needs special handling
	var metascore int
	if (len(line[11]) == 0){
		metascore = 0
	}else{
		var err1 error
		// convert metascore from string to int
		metascore, err1 = strconv.Atoi(line[11])

		if (err1!= nil){
			log.WithFields(log.Fields{"Metascore conversion failed":line[11]}).Info()
			return movie,err1
		}
	}

	movie.Rank = rank
	movie.Title =  line[1]
	movie.Genre =  genrelist
	movie.Description =  line[3]
	movie.Director =  line[4]
	movie.Actors =  line[5]
	movie.Year =  year
	movie.RuntimeMin =  runtime
	movie.Rating =  rating
	movie.Votes =  votes
	movie.RevenueMil =  revenue
	movie.Metascore =  metascore

	return movie,err
}

/******************************************************************************************
 *
 * Get Movies by Year, Year Range and Genre
 *
******************************************************************************************/
func GetMovies(w http.ResponseWriter, r *http.Request) {
	qparams :=  r.URL.Query()

	genre := ""
	year := conf.Settings.DefaultYear

	// validate for query params first

	if qparams["genre"] != nil {
		if len(qparams["genre"][0]) == 0{
			respondWithErrorCode(w, ERR_GENRE_INVALID)
			return
		}
		genre = strings.ToLower(qparams["genre"][0])
	}

	// if both year and year range are provided, return an error
	if (qparams["year"] != nil && (qparams["year_from"] != nil || qparams["year_to"] != nil)) {
        respondWithErrorCode(w, ERR_YEAR_AND_RANGE)
		return
	}

	// if no year query parameters are provided fallback to default year
	if (qparams["year"] == nil && qparams["year_from"] == nil && qparams["year_to"] == nil) {
		movies, err := dao.FindByYear(year, genre)
		if err != nil || movies == nil || len(movies) == 0 {
				log.Info("Responding with No Content")
				respondWithErrorCode(w, ERR_NO_CONTENT)
				return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(movies)
		return
	}

	// Handle if query param is year
	if qparams["year"] != nil{
		year, err := IsValidYear(qparams["year"][0])
		if err != nil {
			respondWithErrorCode(w, ERR_YEAR_INVALID)
			return
		}else{
			movies, err := dao.FindByYear(year, genre)
			if err != nil || movies == nil || len(movies) == 0 {
					log.Info("Responding with No Content")
					respondWithErrorCode(w, ERR_NO_CONTENT)
					return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(movies)
			return
		}

	}

	// Handle if query param is year range
	if qparams["year_from"] != nil && qparams["year_to"] != nil {
		year_from, err := IsValidYear(qparams["year_from"][0])
		if err != nil {
			respondWithErrorCode(w, ERR_YEAR_INVALID)
			return
		}
		year_to, err := IsValidYear(qparams["year_to"][0])
		if err != nil {
			respondWithErrorCode(w, ERR_YEAR_INVALID)
			return
		}

		if year_from > year_to {
			respondWithErrorCode(w, ERR_YEAR_RANGE_INVALID)
			return
		}
		movies, err := dao.FindByYearRange(year_from, year_to, genre)
		if err != nil || movies == nil || len(movies) == 0 {
				respondWithErrorCode(w, ERR_NO_CONTENT)
				return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(movies)
		return

	}else{
		respondWithErrorCode(w, ERR_YEAR_RANGE_INVALID)
		return
	}

}

/******************************************************************************************
 *
 * Validate Year
 *
******************************************************************************************/
func IsValidYear(year string) (int,error){
	if len(year) != 4  {
		err := errors.New("Year is invalid length")
		return -1, err
	}

	iyear, err := strconv.Atoi(year)
	if err != nil {
		return -1, err
	}
	return iyear,err
}
