package main

import (
	"log"
	"net/http"
	"os"

	"github.com/alphatroya/ci-redmine-bindings/settings"
	"github.com/go-redis/redis"
)

func main() {
	settings, err := settings.Current()
	if err != nil {
		log.Fatalf("Failed to create settings %s", err)
	}

	stamper, err := createStamper(settings)
	if err != nil {
		log.Fatalf("Failed to create v2 handler %s", err)
	}
	http.Handle("/bitrise", stamper)
	http.Handle("/bitrise/v2", stamper)
	//nolint
	if err := http.ListenAndServe(":"+settings.Port, nil); err != nil {
		log.Fatal(err)
	}
}

func createStamper(settings *settings.Config) (*Stamper, error) {
	options, err := redis.ParseURL(settings.RedisURL)
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
