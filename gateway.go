package main

import (
	"encoding/json"
	"net/http"

	"./services"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

//Gateway to
type Gateway struct {
	db        *sqlx.DB
	cacheConn services.CacheConnection
}

//Medallion is data structure of the cab medallion
type Medallion struct {
	Id   string `json:"id"`
	Date string `json:"date"`
}

//RequestInfo is data structure of the trip request
type RequestInfo struct {
	Medallions []string `json:"medallions"`
	Date       string   `json:"date"`  //"2014-11-12"
	Fresh      bool     `json:"fresh"` // use cached data or not
}

//MedallionTrip is data structure of the trip list for a particular medallion
type MedallionTrip struct {
	Medallion string          `json:"medallion"`
	Trips     []services.Trip `json:"trips"` //"2014-11-12"
}

func setupGateway(dbDriver, dbSource, hostAddress string) (*gin.Engine, error) {
	// Start and run the server
	db, err := services.InitCabDb(dbDriver, dbSource)
	if err != nil {
		return nil, err
	}

	//init cache
	conn, err := services.InitCache()
	if err != nil {
		return nil, err
	}
	gateway := &Gateway{
		db:        db,
		cacheConn: conn,
	}
	var router *gin.Engine

	//bypass gin.logger
	router = gin.New()
	router.Use(gin.Recovery())

	//gin.SetMode(gin.ReleaseMode)

	// Setup route group for the API
	api := router.Group("api")

	api.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	api.OPTIONS("/", func(c *gin.Context) {
		c.Header("Allow", "POST, GET, OPTIONS")
		c.Header("Content-Type", "application/json")
		c.Status(http.StatusOK)
	})

	api.GET("/"+medallionURL, gateway.medallionHandler())
	api.POST("/"+tripsURL, gateway.TripHandler())
	api.DELETE("/"+cacheURL, gateway.deleteCache())

	go func() {
		// Start and run the server++++`r`
		router.Run(hostAddress)
	}()
	return router, nil
}

// medallionHandler returns a handler func which gets the cab medallion info
func (g *Gateway) medallionHandler() gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		errorMsg := func(err error) {
			c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		}
		if identities, err := services.GetMedallions(g.db); err == nil {
			medallions := []Medallion{}
			for _, i := range identities {
				medallions = append(medallions, Medallion{Id: i.Medallion, Date: i.PickupDatetime.Format("2006-01-02")})
			}
			c.JSON(http.StatusOK, medallions)
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

		if c.Request.Method == "POST" {
			var request = RequestInfo{}
			if err := c.BindJSON(&request); err == nil {
				tripList := []MedallionTrip{}
				for _, m := range request.Medallions {
					if !request.Fresh {
						if data, err := services.GetCache(g.cacheConn, m+request.Date); err == nil {
							trips := []services.Trip{}
							if err = json.Unmarshal([]byte(data), &trips); err == nil {
								tripList = append(tripList, MedallionTrip{Medallion: m, Trips: trips})
								continue
							}
						}
					}
					if trips, err := services.GetTrips(g.db, m, request.Date); err == nil {
						tripList = append(tripList, MedallionTrip{Medallion: m, Trips: trips})
						if data, err := json.Marshal(trips); err == nil {
							services.SetCache(g.cacheConn, m+request.Date, data)
						}
					} else {
						errorMsg(err)
						return
					}
				}
				c.JSON(http.StatusOK, tripList)
			} else {
				errorMsg(err)
			}
		}
	}
}

// deleteCache returns a handler func which clean up cache
func (g *Gateway) deleteCache() gin.HandlerFunc {

	return func(c *gin.Context) {
		services.RemoveCache(g.cacheConn)
		c.Status(http.StatusNoContent)
		return
	}
}
