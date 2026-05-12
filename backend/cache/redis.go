package cache

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func Connect() *redis.Client {
	host := getEnv("REDIS_HOST", "localhost")
	port := getEnv("REDIS_PORT", "6379")
	password := os.Getenv("REDIS_PASSWORD")

	rdb := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0,
	})

	var err error
	for i := 0; i < 30; i++ {
		err = rdb.Ping(Ctx).Err()
		if err == nil {
			break
		}
		log.Printf("Redis not ready (attempt %d/30): %v", i+1, err)
		time.Sleep(time.Second)
	}

	if err != nil {
		log.Fatalf("Could not connect to Redis after 30 attempts: %v", err)
	}

	log.Println("Connected to Redis")
	return rdb
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
