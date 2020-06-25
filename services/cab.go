package services

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

//Identity is data structure of the cab medallion
type Identity struct {
	Medallion      string    `db:"medallion"`
	PickupDatetime time.Time `db:"pickup_datetime"`
}

//Trip is data structure of the cab trip
type Trip struct {
	Medallion       string    `db:"medallion"`
	HackLicense     string    `db:"hack_license"`
	VendorID        string    `db:"vendor_id"`
	RateCode        int       `db:"rate_code"`
	StoreAndFwdFlag string    `db:"store_and_fwd_flag"`
	PickupDatetime  time.Time `db:"pickup_datetime"`
	DropoffDatetime time.Time `db:"dropoff_datetime"`
	PassengerCount  int       `db:"passenger_count"`
	TripTimeInSecs  int       `db:"trip_time_in_secs"`
	TripDistance    float32   `db:"trip_distance"`
}

const (
	queryMedallions = "SELECT DISTINCT medallion, pickup_datetime FROM cab_trip_data"
	queryTrips      = `SELECT medallion, hack_license, vendor_id, rate_code, store_and_fwd_flag,pickup_datetime, dropoff_datetime, passenger_count, trip_time_in_secs,trip_distance 
	FROM cab_trip_data WHERE (medallion = $1 AND pickup_datetime >= $2 AND pickup_datetime < $3)`
)

var GetTripsFunc = func(db *sqlx.DB, medallion string, date string) ([]Trip, error) {

	trips := []Trip{}
	//"2014-11-12"
	parts := strings.Split(date, "-")
	if len(parts) == 3 {
		year, erry := strconv.Atoi(parts[0])
		month, errm := strconv.Atoi(parts[1])
		day, errd := strconv.Atoi(parts[2])
		if erry == nil && errm == nil && errd == nil {
			start := time.Date(year, (time.Month)(month), day, 0, 0, 0, 0, time.UTC)
			end := start.AddDate(0, 0, 1)

			err := db.Select(&trips, queryTrips, medallion, start.Format("2006-01-02"), end.Format("2006-01-02"))
			if err != nil {
				return trips, err
			}
		} else {
			return trips, errors.New("wrong date which needs to be in format of '2014-11-01'")
		}
	} else {
		return trips, errors.New("wrong date which needs to be in format of '2014-11-01'")
	}
	return trips, nil
}

//InitCabDb to initialize Cab DB connection
func InitCabDb(dbDriver, dbSource string) (*sqlx.DB, error) {
	// Start and run the server
	return sqlx.Connect(dbDriver, dbSource)
}

//GetMedallions to get medallion list
func GetMedallions(db *sqlx.DB) ([]Identity, error) {
	medallions := []Identity{}
	err := db.Select(&medallions, queryMedallions)
	return medallions, err
}

//GetTrips to get a list of trip details for a list of medallions in a particular day
func GetTrips(db *sqlx.DB, medallion string, date string) ([]Trip, error) {
	return GetTripsFunc(db, medallion, date)
}
