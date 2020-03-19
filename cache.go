package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

// GetCache returns a handler func which retrieves data from cache
func (g *Gateway) GetCache() gin.HandlerFunc {

	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			var addrs = strings.Split(c.Request.URL.Path, "/")
			req := strings.ToLower(addrs[len(addrs)-1])
			queries := c.Request.URL.Query()
			if _, ok := queries["ignore"]; !ok {
				if req == medallionURL {
					if data, err := g.getCache(req); err == nil {
						ids := []string{}
						err = json.Unmarshal([]byte(data), &ids)
						if err == nil {
							c.AbortWithStatusJSON(http.StatusOK, ids)
							return
						}
					}
				} else if req == searchURL {
					if ids, ok := queries["medallion"]; ok {
						tripArry := []Trip{}
						for _, id := range ids {
							if dates, ok := queries["date"]; ok {

								for _, date := range dates {
									if data, err := g.getCache(id + date); err == nil {
										trips := []Trip{}
										err = json.Unmarshal([]byte(data), &trips)
										if err == nil {
											for _, t := range trips {
												tripArry = append(tripArry, t)
											}
										} else {
											goto Next
										}
									} else {
										goto Next
									}
								}
							} else {
								if data, err := g.getCache(id); err == nil {
									trips := []Trip{}
									err = json.Unmarshal([]byte(data), &trips)
									if err == nil {
										for _, t := range trips {
											tripArry = append(tripArry, t)
										}
									} else {
										goto Next
									}
								} else {
									goto Next
								}
							}
						}
						c.AbortWithStatusJSON(http.StatusOK, tripArry)
						return
					}
				}
			}

		}
	Next:
		c.Next()
	}
}

// GetCache returns a handler func which retrieves data from cache
func (g *Gateway) clearCache() gin.HandlerFunc {

	return func(c *gin.Context) {
		if c.Request.Method == "PUT" {
			err := g.removeCache()
			c.JSON(http.StatusOK, err)
			return
		}
		c.Header("Allow", http.MethodPut)
		c.Status(http.StatusBadRequest)
	}
}

func newPool() *redis.Pool {
	return &redis.Pool{
		// Maximum number of idle connections in the pool.
		MaxIdle: 80,
		// max number of connections
		MaxActive: 12000,
		// Dial is an application supplied function for creating and
		// configuring a connection.
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

func initRedis() (redis.Conn, error) {
	pool := newPool()
	conn := pool.Get()
	// Send PING command to Redis
	_, err := conn.Do("PING")
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (g *Gateway) closeRedis() {
	g.redisConn.Close()
}

// setCache executes the redis SET command
func (g *Gateway) setCache(key string, value []byte) error {
	reply, err := g.redisConn.Do("SET", key, value)
	if err != nil {
		fmt.Printf("reply %v\n", reply)
		return err
	}
	return nil
}

// getCache executes the redis GET command
func (g *Gateway) getCache(key string) (string, error) {
	s, err := redis.String(g.redisConn.Do("GET", key))
	if err == redis.ErrNil {
		return "", errors.New(fmt.Sprintf("%s does not exist\n", key))
	} else if err != nil {
		return "", err
	}
	return s, nil
}

func (g *Gateway) removeCache() error {
	reply, err := g.redisConn.Do("FLUSHDB")
	if err != nil {
		fmt.Printf("reply %v\n", reply)
		return err
	}
	return nil
}
