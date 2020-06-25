package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCabQuery(t *testing.T) {
	const id = "11B683D81211E96F472173C69D4632D6"
	const date = "2013-12-31"
	db, err := InitCabDb("sqlite3", "cab.db")
	assert.Equal(t, err, nil)

	trips, err := GetTrips(db, id, date)
	assert.Equal(t, err, nil)
	assert.Equal(t, 29, len(trips))
	for _, trip := range trips {
		assert.Equal(t, id, trip.Medallion)
		assert.Equal(t, date, trip.PickupDatetime.Format("2006-01-02"))
	}
}
