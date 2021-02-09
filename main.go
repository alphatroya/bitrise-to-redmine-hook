package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis"
)

const (
	redisURLEnvKey = "REDIS_URL"
)

func main() {
	v2, err := createStamper()
	if err != nil {
		log.Fatalf("failed to create v2 handler %s", err)
	}
	http.Handle("/bitrise", &Handler{&EnvSettingsBuilder{}})
	http.Handle("/bitrise/v2", v2)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func createStamper() (*Stamper, error) {
	redisURL := os.Getenv(redisURLEnvKey)
	if len(redisURL) == 0 {
		return nil, fmt.Errorf("%s should be set", redisURLEnvKey)
	}
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(options)
	_, err = rdb.Ping().Result()
	if err != nil {
		return nil, err
	}
	return NewStamper(&EnvSettingsBuilder{}, rdb), nil
}
