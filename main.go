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
	sb := &EnvSettingsBuilder{}
	settings, err := sb.build()
	if err != nil {
		log.Fatalf("Failed to create settings %s", err)
	}

	v2, err := createStamper(settings)
	if err != nil {
		log.Fatalf("Failed to create v2 handler %s", err)
	}
	http.Handle("/bitrise", &Handler{sb})
	http.Handle("/bitrise/v2", v2)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func createStamper(settings *Settings) (*Stamper, error) {
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

	return NewStamper(settings, rdb), nil
}
