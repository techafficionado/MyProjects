# REST Service for IMDB Movies

The IMDB Movies REST API Service allows user to upload movies list from IMDB in a specific format and query the movie list by Year and Genre. The uploaded CSV file is parsed and saved to the database. Based on filter/query parameters provided , the top 10 popular movies are returned sorted by the rating of the movie from highest to lowest.

The application runs on port 8000 on the localhost.
# Project

The code is written using [Go](https://golang.org/doc/), a highly efficient and concurrent language and the backend database is MongoDB.


## Structure
All the Go Files are in the Main directory(imdb):
* main.go
* rest.go
* version.go
* model.go
* rest_test.go

All the Data files are in the data subdirectory(imdb/data):
* config.toml
* swagger.yaml

The application binary is in the main directory(imdb):
* imdb-restapi

The Dockerfile is in the main directory(imdb):
* Dockerfile

The Docker tar image is in the main directory(imdb):
* imdb-restapi-docker.tar

In addition, the following csv files are present in the 'test' sub directory(imdb/test) for testing purposes:
* IMDB-Movie-Data_Assignment.csv
* largefile.csv
* faillist.csv
* passlist.csv

logs directory is in the main directory(imdb):
* logs/imdb-restapi.log

## Endpoints
Please refer to swagger.yaml for a detailed description

* POST http://localhost:8000/imdb/uploadmovies 
Upload a multipart/form-data CSV file with keyname as 'file'
Note: This has been consciously named to 'uploadmovies' to signify that there is a file upload here. This could very well have been named just 'movies'

* http://localhost:8000/imdb/movies
Get movies by year/year-range and genre

* http://localhost:8000/imdb/version
Get Version of the Application

* http://localhost:8000/imdb/endpoints
Get swagger.yaml on the endpoints

## Building

A proper docker image has been provided. Please refer to the Deployment section.
However, If the intention is to verify the project build please proceed further.

In order for the project to build properly, the $GOPATH is expected to be setup properly. The source code of the project should be a sub-directory of $GOPATH/src.

You can build the whole project using the sub-path and the directory name, if the source is stored within the GOPATH (as it should). The below build command generates a binary with the name 'imdb-restapi', within the directory where the command is executed.

For all the additional Go libraries that were used in the project please do a 'go get' for all the libraries listed under the 'Go Libraries' section.

```
go build -o imdb-restapi

```

A Docker image could be built as shown

```
sudo docker build -t imdb-restapi-docker

```

## Deployment

### Docker

The deployment through the Docker image that has been provided with this package requires: 
* Docker installation
* MongoDB installation(No configuration is needed). The database and indexes(title,year) will be configured through the Application. 
Note:The expectation is that the MongoDB is on the same localhost as the application
* Make sure MongoDB is running using:
sudo systemctl status mongod

Once the above installations and verifications are complete, run the docker image as shown:

* Load the tar
sudo docker load -i imdb-restapi-docker.tar
sudo docker run --net=host --name imdb-restapi-v1 -d imdb-restapi-docker
349cd1ba4be9c86e1e0263b9a479b2272b53f5c38ff15d1198bf22ec21d10f21
$ sudo docker ps -a
CONTAINER ID        IMAGE               COMMAND               CREATED              STATUS                          PORTS               NAMES
3aed8d73b64e        imdb-restapi-docker   "/app/imdb-restapi"   5 seconds ago        Up 4 seconds                                        imdb-restapi-v1

sudo docker run --net=host -it imdb-restapi-docker (optional alternative)

### Manual

The deployment of the project is straighforward. For the project to run, it requires:
* a configuration file (data/config.toml)
* a binary of application (imdb-restapi)
* MongoDB installation(No configuration is needed). The database and indexes will be configured through the Application
Note:The expectation is that the MongoDB is on the same localhost as the application
* Make sure MongoDB is running using:
sudo systemctl status mongod

To run the application:
./imdb-restapi

Once the Application is up and running, please refer to the 'swagger.yaml' document to start using the API.

Alternatively you can query the endpoints and version APIs as shown:



## IMDB Movies REST API Help Guide
* POST Request to Upload Movies
Request:
curl -F file=@IMDB-Movie-Data_Assignment.csv http://localhost:8000/imdb/uploadmovies
Response:
{"RecordsRead":1000,"RecordsCreated":1000,"RecordsErrored":0}
Note: A unique composite index on Title,Year is created when the Application starts up. So if duplicate Records are uploaded, you would see them 'RecordsErrored'

* GET:
Request:
curl -F file=@largefile.csv http://localhost:8000/imdb/uploadmovies
Response:
{"code":"400","error":"File is too large. Maximum upload size is 2097152 Bytes"}

Request:
$ curl -XGET "http://localhost:8000/imdb/movies?year=20161" | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100    52  100    52    0     0   5425      0 --:--:-- --:--:-- --:--:--  5777
Response:
{
  "code": "400",
  "error": "Please provide a valid year"
}

Request:
$ curl -XGET "http://localhost:8000/imdb/movies?year=2016&year_from=2016" | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100    79  100    79    0     0   8240      0 --:--:-- --:--:-- --:--:--  8777
Response:
{
  "code": "400",
  "error": "Please provide either the year or a range but not both"
}

Request:
$ curl -XGET "http://localhost:8000/imdb/movies?year_to=2015&year_from=2016" | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100    92  100    92    0     0   9654      0 --:--:-- --:--:-- --:--:-- 10222
Response:
{
  "code": "400",
  "error": "Please provide a valid year_from and year_to in chronological order"
}

Request:
$ curl -XGET "http://localhost:8000/imdb/movies?year_to=205&year_from=2016" | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100    52  100    52    0     0   5446      0 --:--:-- --:--:-- --:--:--  5777
Response:
{
  "code": "400",
  "error": "Please provide a valid year"
}

Request:
$ curl -XGET "http://localhost:8000/imdb/movies?year=2016&genre=" | jq         % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100    53  100    53    0     0   5536      0 --:--:-- --:--:-- --:--:--  5888
Response:
{
  "code": "400",
  "error": "Please provide a valid genre"
}

Request:
$ curl -XGET "http://localhost:8000/imdb/movies?year=2015&genre=war" | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   697  100   697    0     0  63237      0 --:--:-- --:--:-- --:--:-- 69700
Response:
[
  {
    "title": "Eye in the Sky",
    "genre": [
      "drama",
      "thriller",
      "war"
    ],
    "description": "Col. Katherine Powell, a military officer in command of an operation to capture terrorists in Kenya, sees her mission escalate when a girl enters the kill zone triggering an international dispute over the implications of modern warfare.",
    "year": 2015,
    "runtime_min": 102,
    "rating": 7.3
  },
  {
    "title": "Macbeth",
    "genre": [
      "drama",
      "war"
    ],
    "description": "Macbeth, the Thane of Glamis, receives a prophecy from a trio of witches that one day he will become King of Scotland. Consumed by ambition and spurred to action by his wife, Macbeth murders his king and takes the throne for himself.",
    "year": 2015,
    "runtime_min": 113,
    "rating": 6.7
  }
]


## Testing

Unit tests are provided with the project defined in file rest_test.go. They can be run using the below command (standard Go command), within the project's directory:

```
go test -v .
=== RUN   TestPostCSVContentType
--- PASS: TestPostCSVContentType (0.00s)
=== RUN   TestPostCSVFileSize
--- PASS: TestPostCSVFileSize (0.01s)
=== RUN   TestPostCSVCreate
--- PASS: TestPostCSVCreate (0.00s)
=== RUN   TestPostCSVDuplicate
--- PASS: TestPostCSVDuplicate (0.00s)
=== RUN   TestPostCSVFormat
--- PASS: TestPostCSVFormat (0.00s)
=== RUN   TestGetByInvalidYear
--- PASS: TestGetByInvalidYear (0.00s)
=== RUN   TestGetByInvalidGenre
--- PASS: TestGetByInvalidGenre (0.00s)
=== RUN   TestGetByYearAndYearRange
--- PASS: TestGetByYearAndYearRange (0.00s)
=== RUN   TestGetByChronoOrder
--- PASS: TestGetByChronoOrder (0.00s)
PASS
ok      imdb   0.030s

```

All the tests are expected to pass.

## Benchmarks

No benchmarking has been performed on this code.

## Go Libraries

The following libraries have been used in the project:
* [github.com/sirupsen/logrus](https://github.com/sirupsen/logrus)
* [github.com/gorilla/mux](https://github.com/gorilla/mux)
* [github.com/BurntSushi/toml](https://github.com/BurntSushi/toml)
* [gopkg.in/mgo.v2](https://godoc.org/gopkg.in/mgo.v2)
* [net/http](https://golang.org/pkg/net/http/)
* [encoding/csv](https://golang.org/pkg/encoding/csv/)
* [encoding/json](https://golang.org/pkg/encoding/json/)
* [net/http/httptest](https://golang.org/pkg/net/http/httptest/)


## Learn about Go

The below are good links to get started with Go:

* [Go](https://golang.org/doc/)
* [Go wiki](https://github.com/golang/go/wiki)
* [A getting started guide for Go newcomes](https://github.com/alco/gostart)


## MongoDB Installation

* Step 1: Import the public key used by the package management system.

sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 2930ADAE8CAF5059EE73BB4B58712A2291FA4AD5

* Step 2: Create a list file for MongoDB. (for Ubuntu 16.04)
vi /etc/apt/sources.list.d/mongodb-org-3.6.list
	
Add 'deb http://repo.mongodb.org/apt/ubuntu xenial/mongodb-org/3.6 multiverse' and save

* Step 3: sudo apt-get update

* Step 4: sudo apt-get install -y mongodb-org
Make sure you have /data/db with right permissions as Mongo uses it.

* Step 5: Start the MongoDB service 
sudo systemctl start mongod		

* Step 6: Check the status of MongoDB service 
sudo systemctl status mongod

* Step 7: Verification
'mongo' to get to the prompt 

