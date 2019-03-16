/******************************************************************************
 * \file        main.go
 *
 * \brief       Main GO File for IMDB Movies REST Service
 *
 * \author      Reshma Syeda
 *
 * ****************************************************************************/

package main

import (
    log "github.com/sirupsen/logrus"
    "net/http"
    "github.com/gorilla/mux"
    "os"
    "fmt"
    "io"
    "github.com/BurntSushi/toml"
)

// TomlConfig struct for config file
type TomlConfig struct {
    App struct {
        Port string `toml:"port"`
        Logdir string `toml:"logdir"`
    } `toml:"app"`
    Database struct {
        Server string `toml:"server"`
        Port string `toml:"port"`
        DBName string `toml:"dbname"`
    } `toml:"database"`
	Settings struct{
		DefaultYear int `toml:"defaultyear"`
		FileSizeKB int64 `toml:"filesizekb"`
	}
}

// Config File
var conf TomlConfig

// Database Access Object
var dao = MoviesDAO{}

/******************************************************************************************
 *
 * Initialize Logger to log to File and Console
 *
*******************************************************************************************/
func InitLogger() {
    var filename string = "imdb-restapi.log"
    filename = conf.App.Logdir + filename
    log.Info(filename)
    f, err := os.OpenFile(filename, os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0644)
    Formatter := new(log.TextFormatter)
    Formatter.FullTimestamp = true
    log.SetFormatter(Formatter)

    mw := io.MultiWriter(os.Stdout, f)
    if err != nil {
        // Cannot open log file. Logging to stderr
        fmt.Println(err)
    }else{
        log.SetOutput(mw)
    }
    log.Info("Initialized Logger")
}


/******************************************************************************************
 *
 * Main Function
 *
*******************************************************************************************/
func main() {

    if _, err := toml.DecodeFile("data/config.toml", &conf); err != nil {
        log.Fatal(err)
    }

    InitLogger()

    log.WithFields(log.Fields{"Application Port":conf.App.Port}).Info()
    log.WithFields(log.Fields{"Database Port":conf.Database.Port}).Info()
    log.WithFields(log.Fields{"Database Name":conf.Database.DBName}).Info()
    log.WithFields(log.Fields{"Max File Size KB":conf.Settings.FileSizeKB}).Info()

	dao.Server = conf.Database.Server
	dao.Database =  conf.Database.DBName
	dao.Connect()

	log.WithFields(log.Fields{"Established connection to database":dao.Database}).Info()

    router := mux.NewRouter()
    router.HandleFunc("/imdb/version", GetVersion).Methods("GET") // get version
    router.HandleFunc("/imdb/uploadmovies", PostCSV).Methods("POST") // post movie uploads
    router.HandleFunc("/imdb/movies", GetMovies).Methods("GET") // get movies
    router.HandleFunc("/imdb/endpoints", GetEndpoints).Methods("GET") // get Rest Endpoint Info

	log.Info("Server is up and ready")
    log.Fatal(http.ListenAndServe(conf.App.Port, router))
}
