package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

//Gateway to
type Gateway struct {
	db        *sqlx.DB
	redisConn redis.Conn
}

const (
	medallionURL  = "medallion"
	searchURL     = "search"
	clearCacheURL = "clear"
)

var (
	dbDriver    string
	dbSource    string
	hostAddress string
)

func init() {
	flag.StringVar(&dbDriver, "dbDriver", "sqlite3", "the database driver name")
	flag.StringVar(&dbSource, "dbSource", "cab.db", "the database data source name")
	flag.StringVar(&hostAddress, "hostAddress", "0.0.0.0:3000", "the host address")
}

func main() {

	flag.Parse()

	signalled := make(chan os.Signal)
	signal.Notify(signalled, os.Interrupt)
	signal.Notify(signalled, os.Kill)
	signal.Notify(signalled, syscall.SIGTERM)

	// Start and run the server
	db, err := sqlx.Connect(dbDriver, dbSource)
	if err != nil {
		fmt.Print("Error: ", err.Error())
		os.Exit(1)
	}

	//init redis
	conn, err := initRedis()
	if err != nil {
		fmt.Print("Error: ", err.Error())
		os.Exit(2)
	}
	gateway := &Gateway{
		db:        db,
		redisConn: conn,
	}

	//gin.SetMode(gin.ReleaseMode)

	var router *gin.Engine

	//bypass gin.logger
	router = gin.New()
	router.Use(gin.Recovery())

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

	api.GET("/"+medallionURL, gateway.GetCache(), gateway.CabHandler())
	api.GET("/"+searchURL, gateway.GetCache(), gateway.TripHandler())
	api.Any("/"+clearCacheURL, gateway.clearCache())

	go func() {
		// Start and run the server
		router.Run(hostAddress)
	}()

	s := <-signalled

	gateway.closeRedis()

	fmt.Printf("Received %s signal. Quitting...\n", s)
}
