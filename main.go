package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const (
	medallionURL = "medallion"
	tripsURL     = "trips"
	cacheURL     = "cache"
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

	_, err := setupGateway(dbDriver, dbSource, hostAddress)
	if err != nil {
		fmt.Print("Error: ", err.Error())
		os.Exit(1)
	}
	s := <-signalled
	fmt.Printf("Received %s signal. Quitting...\n", s)
}
