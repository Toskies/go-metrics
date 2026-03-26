package main

import (
	"log"
	"net/http"

	"github.com/Toskies/go-metrics/internal/oteldemo"
)

func main() {
	handler, err := oteldemo.NewHandler()
	if err != nil {
		log.Fatal(err)
	}

	addr := ":2113"
	log.Printf("oteldemo listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
