/******************************************************************************
 * \file        model.go
 *
 * \brief       GO File that has Database Model and Access Object defined
 *
 * \author      Reshma Syeda
 *
 * ****************************************************************************/

package main

import (
    log "github.com/sirupsen/logrus"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

// Database Access Object
type MoviesDAO struct {
	Server   string
	Database string
}

var db *mgo.Database

const (
	COLLECTION = "movies"
)

/******************************************************************************************
 *
 * Establish a connection to database
 *
*******************************************************************************************/
func (m *MoviesDAO) Connect() {
	session, err := mgo.Dial(m.Server)
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB(m.Database)

	// Add Unique Composite Index for Title and Year
	index := mgo.Index{
		Key: []string{"title", "year"},
		Unique: true,
		DropDups: true,
		Background: true,
		Sparse: true,
	}
	db.C(COLLECTION).EnsureIndex(index)
}

/******************************************************************************************
 *
 * Find list of movies by specific year and genre
 *
*******************************************************************************************/
func (m *MoviesDAO) FindByYear(year int, genre string) ([]MovieGet, error) {
	var movies []MovieGet
	if len(genre) == 0 {
		err := db.C(COLLECTION).Find(bson.M{"year":year}).
								Sort("-rating").
								Limit(10).
								All(&movies)
		return movies, err
	}

	err := db.C(COLLECTION).Find(bson.M{"year":year, "genre": bson.M{"$eq":genre}}).
							Sort("-rating").
							Limit(10).
							All(&movies)
	return movies, err
}

/******************************************************************************************
 *
 * Find list of movies by an year range and genre
 *
*******************************************************************************************/
func (m *MoviesDAO) FindByYearRange(yearfrom int, yearto int, genre string) ([]MovieGet, error) {
	var movies []MovieGet
	if len(genre) == 0 {
		err := db.C(COLLECTION).Find(bson.M{"year":bson.M{"$gte":yearfrom,"$lte":yearto}}).
									Sort("-rating").
									Limit(10).
									All(&movies)
		return movies, err
	}

	err := db.C(COLLECTION).Find(bson.M{"year":bson.M{"$gte":yearfrom,"$lte":yearto},
										"genre":bson.M{"$eq":genre}}).
							Sort("-rating").
							Limit(10).
							All(&movies)
	return movies, err

}


/******************************************************************************************
 *
 * Insert a movie into database
 *
*******************************************************************************************/
func (m *MoviesDAO) Insert(movie Movie) error {
	err := db.C(COLLECTION).Insert(&movie)
	return err
}

/******************************************************************************************
 *
 * Clean the database
 *
*******************************************************************************************/
func (m *MoviesDAO) Clean() error {
	log.Warning("Cleaning Database!!!!!!!!!!!!!")
	_, err := db.C(COLLECTION).RemoveAll(bson.M{})
	return err
}
