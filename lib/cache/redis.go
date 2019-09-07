package cache

import (
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"os"
)

// RedisClient represents redis client.
var RedisClient *redis.Client

// ConnectRedis connects redis.
func ConnectRedis() *redis.Client {

	hostname := os.Getenv("REDIS_HOSTNAME")
	password := os.Getenv("REDIS_PASSWORD")

	client := redis.NewClient(&redis.Options{
		Addr:     hostname,
		Password: password,
		DB:       0,
	})

	pong, err := client.Ping().Result()
	if err != nil {
		log.Fatalf("Redis client error: %s", err)
	}

	log.Info(pong)
	RedisClient = client
	return client
}
