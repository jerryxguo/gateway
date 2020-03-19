package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

//Identity is data structure of the cab medallion
type Identity struct {
	Medallion string `db:"medallion"`
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
	queryTripsByDate = `SELECT medallion, hack_license, vendor_id, rate_code, store_and_fwd_flag,pickup_datetime, dropoff_datetime, passenger_count, trip_time_in_secs,trip_distance 
	FROM cab_trip_data WHERE (medallion = $1 AND pickup_datetime >= $2 AND pickup_datetime < $3)`
	queryMedallions       = "SELECT DISTINCT Medallion FROM cab_trip_data"
	queryTripsByMedallion = `SELECT medallion, hack_license, vendor_id, rate_code, store_and_fwd_flag,pickup_datetime, dropoff_datetime, passenger_count, trip_time_in_secs,trip_distance 
	FROM cab_trip_data WHERE medallion = $1`
)

// CabHandler returns a handler func which gets the cab id info
func (g *Gateway) CabHandler() gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		errorMsg := func(err error) {
			c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		}
		if identities, err := g.getMedallions(); err == nil {
			ids := []string{}
			for _, i := range identities {
				ids = append(ids, i.Medallion)
			}
			c.JSON(http.StatusOK, ids)

			var addrs = strings.Split(c.Request.URL.Path, "/")
			req := strings.ToLower(addrs[len(addrs)-1])
			if data, err := json.Marshal(ids); err == nil {
				g.setCache(req, data)
			}
		} else {
			errorMsg(err)
		}

	}
}

// TripHandler returns a handler func which gets the trip info
func (g *Gateway) TripHandler() gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		errorMsg := func(err error) {
			c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		}

		if c.Request.Method == "GET" {
			queries := c.Request.URL.Query()
			fmt.Printf("%v\n", queries)
			if ids, ok := queries["medallion"]; ok {
				tripArry := []Trip{}
				for _, id := range ids {
					if dates, ok := queries["date"]; ok {
						//"2014-11-12"
						for _, date := range dates {
							parts := strings.Split(date, "-")
							if len(parts) == 3 {
								year, erry := strconv.Atoi(parts[0])
								month, errm := strconv.Atoi(parts[1])
								day, errd := strconv.Atoi(parts[2])
								if erry == nil && errm == nil && errd == nil {
									start := time.Date(year, (time.Month)(month), day, 0, 0, 0, 0, time.UTC)
									end := start.AddDate(0, 0, 1)
									trips, err := g.getTripsByDate(id, start.Format("2006-01-02"), end.Format("2006-01-02"))
									if err != nil {
										errorMsg(err)
										return
									}
									for _, t := range trips {
										tripArry = append(tripArry, t)
									}
									if data, err := json.Marshal(trips); err == nil {
										g.setCache(id+date, data)
									}

								} else {
									errorMsg(errors.New("wrong date"))
									return
								}
							} else {
								errorMsg(errors.New("wrong date"))
								return
							}
						}

					} else {
						trips, err := g.getTripsByMedallion(id)
						if err != nil {
							errorMsg(err)
							return
						}
						for _, t := range trips {
							tripArry = append(tripArry, t)
						}
						if data, err := json.Marshal(trips); err == nil {
							g.setCache(id, data)
						}
					}
				}
				c.JSON(http.StatusOK, tripArry)
			} else {
				/* not valid */
				errorMsg(errors.New("no medallion supplied"))
				return
			}

		} else if c.Request.Method == "POST" {
			var request = make(map[string]string)
			if err := c.BindJSON(&request); err == nil {
				if id, ok := request["medallion"]; ok {
					if date, ok := request["date"]; ok {
						//"2014-11-12"
						parts := strings.Split(date, "-")
						if len(parts) == 3 {
							year, erry := strconv.Atoi(parts[0])
							month, errm := strconv.Atoi(parts[1])
							day, errd := strconv.Atoi(parts[2])
							if erry == nil && errm == nil && errd == nil {
								start := time.Date(year, (time.Month)(month), day, 0, 0, 0, 0, time.UTC)
								end := start.AddDate(0, 0, 1)
								trips, err := g.getTripsByDate(id, start.Format("2006-01-02"), end.Format("2006-01-02"))
								if err != nil {
									errorMsg(err)
									return
								}
								c.JSON(http.StatusOK, trips)

							} else {
								errorMsg(errors.New("wrong date"))
								return
							}
						} else {
							errorMsg(errors.New("wrong date"))
							return
						}

					} else {
						/* not valid */
						errorMsg(errors.New("date doesn't exist"))
						return
					}

				} else {
					/* not valid */
					errorMsg(errors.New("medallion doesn't exist"))
					return
				}
			} else {
				errorMsg(err)
			}
		}
	}
}

func (g *Gateway) getMedallions() ([]Identity, error) {
	medallions := []Identity{}
	err := g.db.Select(&medallions, queryMedallions)
	return medallions, err
}

func (g *Gateway) getTripsByDate(medallion, startTime, endTime string) ([]Trip, error) {
	trips := []Trip{}
	err := g.db.Select(&trips, queryTripsByDate, medallion, startTime, endTime)
	if err != nil {
		return nil, err
	}

	return trips, nil
}

func (g *Gateway) getTripsByMedallion(medallion string) ([]Trip, error) {
	trips := []Trip{}
	err := g.db.Select(&trips, queryTripsByMedallion, medallion)
	if err != nil {
		return nil, err
	}
	return trips, nil
}
