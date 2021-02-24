package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis"
)

func main() {
	sb := &EnvSettingsBuilder{}
	settings, err := sb.build()
	if err != nil {
		log.Fatalf("Failed to create settings %s", err)
	}

	stamper, err := createStamper(settings)
	if err != nil {
		log.Fatalf("Failed to create v2 handler %s", err)
	}
	http.Handle("/bitrise", stamper)
	http.Handle("/bitrise/v2", stamper)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func createStamper(settings *Settings) (*Stamper, error) {
	options, err := redis.ParseURL(settings.redisURL)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(options)
	_, err = rdb.Ping().Result()
	if err != nil {
		return nil, err
	}
	logger := log.New(os.Stdout, "Stamper: ", log.LstdFlags)

	return NewStamper(settings, rdb, logger), nil
}
