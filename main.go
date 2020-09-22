package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	http.Handle("/bitrise", &Handler{&EnvSettingsBuilder{}})
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
