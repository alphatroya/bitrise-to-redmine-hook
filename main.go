package main

import (
	"log"
	"net/http"
	"os"
)

const (
	redisURLEnvKey = "REDIS_URL"
)

func main() {
	redis := os.Getenv(redisURLEnvKey)
	if len(redis) == 0 {
		log.Fatalf("%s should be set", redisURLEnvKey)
	}
	http.Handle("/bitrise", &Handler{&EnvSettingsBuilder{}})
	v2, err := NewStamper(&EnvSettingsBuilder{}, redis)
	if err != nil {
		log.Fatalf("failed to create v2 handler")
	}
	http.Handle("/bitrise/v2", v2)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
