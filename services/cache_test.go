package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	const key = "key"
	const data = "[{'data':'whatever'}]"
	//init cache
	conn, err := InitCache()
	assert.Equal(t, err, nil)

	SetCache(conn, key, []byte(data))

	ret, err := GetCache(conn, key)
	assert.Equal(t, err, nil)
	assert.Equal(t, data, ret)
}
