package config

import (
	"fmt"

	"github.com/go-redis/redis"
)

// redis connection object
var RDB *redis.Client

func initRDB() {
	redisHost := Config("REDIS_HOST")
	redisPort := Config("REDIS_PORT")

	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	RDB = redis.NewClient(&redis.Options{
		Addr:     addr, // host:port of the redis server
		Password: "",   // no password set
		DB:       0,    // use default DB
	})

	// testing connection with redis server
	if _, err := RDB.Ping().Result(); err != nil {
		panic(fmt.Errorf("error connecting with redis: %+v", err))
	}
}
