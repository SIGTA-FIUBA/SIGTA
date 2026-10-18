package main

import (
	"net/http"
	"backend/src/adapters/rest"
	"backend/src/application"
)

func main() {
      greeter := &application.GreeterService{}
      api := rest.NewServer(rest.NewGreeterHandler(greeter))

      router := http.NewServeMux()
      rest.HandlerFromMux(rest.NewStrictHandler(api, nil), router)

      server := &http.Server{
              Addr:    ":8080",
              Handler: router,
      }

      server.ListenAndServe()
}
