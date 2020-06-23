package services

import (
	"fmt"

	"github.com/pkg/errors"
)

//CacheConnection is cache connection
type CacheConnection map[string][]byte

//InitCache to set up cache
func InitCache() (CacheConnection, error) {
	cache := make(map[string][]byte)
	return cache, nil
}

// SetCache to set cache
func SetCache(conn CacheConnection, key string, value []byte) {
	conn[key] = value
}

// GetCache to get cache
func GetCache(conn CacheConnection, key string) (string, error) {
	if val, ok := conn[key]; ok {
		return string(val), nil
	}
	return "", errors.New(fmt.Sprintf("No cache exists for %s", key))
}

// RemoveCache to clean cache
func RemoveCache(conn CacheConnection) {
	for key := range conn {
		delete(conn, key)
	}
}
