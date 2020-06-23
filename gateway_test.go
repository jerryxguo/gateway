package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"./services"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

const (
	year   = 2000
	month  = 12
	day    = 1
	first  = "1"
	second = "2"
)

var tripList = []services.Trip{
	services.Trip{
		Medallion:      first,
		PickupDatetime: time.Date(year, month, day, 0, 0, 0, 0, time.UTC),
		PassengerCount: 1,
	},
	services.Trip{
		Medallion:      first,
		PickupDatetime: time.Date(year, month, day, 0, 0, 0, 0, time.UTC),
		PassengerCount: 2,
	},
	services.Trip{
		Medallion:      second,
		PickupDatetime: time.Date(year, month, day, 0, 0, 0, 0, time.UTC),
		PassengerCount: 1,
	},
}

func setup() (*gin.Engine, error) {

	services.InitCabDbFunc = func(dbDriver, dbSource string) (*sqlx.DB, error) { return nil, nil }
	router, err := setupGateway("", "", "0.0.0.0:3000")
	return router, err
}

func TestWrongDate(t *testing.T) {
	router, err := setup()
	assert.Equal(t, err, nil)
	// wrong date format
	searchString := fmt.Sprintf("{\"medallions\": [\"%s\", \"%s\"], \"date\":\"%d-%02d-%02d-10:11:00\"}", first, second, year, month, day)
	req, _ := http.NewRequest("POST", "/api/trips", bytes.NewBuffer([]byte(searchString)))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNormalQuery(t *testing.T) {

	services.GetTripsFunc = func(db *sqlx.DB, medallion string, date string) ([]services.Trip, error) {
		trips := []services.Trip{}
		for _, trip := range tripList {
			if medallion == trip.Medallion {
				trips = append(trips, trip)
			}
		}
		return trips, nil
	}

	router, err := setup()
	assert.Equal(t, err, nil)
	//test normal query
	searchString := fmt.Sprintf("{\"medallions\": [\"%s\", \"%s\"], \"date\":\"%d-%02d-%02d\"}", first, second, year, month, day)
	req, _ := http.NewRequest("POST", "/api/trips", bytes.NewBuffer([]byte(searchString)))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var medallionTrip = []MedallionTrip{}
	err = json.Unmarshal([]byte(w.Body.String()), &medallionTrip)
	assert.Equal(t, 2, len(medallionTrip))
	assert.Nil(t, err)
	for _, m := range medallionTrip {
		if m.Medallion == first {
			assert.Equal(t, 2, len(m.Trips))
		} else if m.Medallion == second {
			assert.Equal(t, 1, len(m.Trips))
		} else {
			assert.True(t, false)
		}
	}

	//test cache policy works
	services.GetTripsFunc = func(db *sqlx.DB, medallion string, date string) ([]services.Trip, error) {
		return []services.Trip{}, nil
	}
	req, _ = http.NewRequest("POST", "/api/trips", bytes.NewBuffer([]byte(searchString)))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal([]byte(w.Body.String()), &medallionTrip)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(medallionTrip))
	assert.Nil(t, err)
	for _, m := range medallionTrip {
		if m.Medallion == first {
			assert.Equal(t, 2, len(m.Trips))
		} else if m.Medallion == second {
			assert.Equal(t, 1, len(m.Trips))
		} else {
			assert.True(t, false)
		}
	}

	// if ignore cache
	searchString = fmt.Sprintf("{\"medallions\": [\"%s\", \"%s\"], \"date\":\"%d-%02d-%02d\",\"fresh\":true}", first, second, year, month, day)
	req, _ = http.NewRequest("POST", "/api/trips", bytes.NewBuffer([]byte(searchString)))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal([]byte(w.Body.String()), &medallionTrip)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(medallionTrip))
	assert.Nil(t, err)
	for _, m := range medallionTrip {
		if m.Medallion == first {
			assert.Equal(t, 0, len(m.Trips))
		} else if m.Medallion == second {
			assert.Equal(t, 0, len(m.Trips))
		} else {
			assert.True(t, false)
		}
	}

}
