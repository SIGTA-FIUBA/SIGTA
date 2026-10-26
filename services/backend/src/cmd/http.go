package main

import (
	"log"
	"net/http"
	"time"

	"backend/src/adapters/rest"
	"backend/src/application"
)

func main() {
	greeter := &application.GreeterService{}
	api := rest.NewServer(rest.NewGreeterHandler(greeter))

	router := http.NewServeMux()
	rest.HandlerFromMux(rest.NewStrictHandler(api, nil), router)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
