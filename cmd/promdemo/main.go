package main

import (
	"log"
	"net/http"

	"github.com/Toskies/go-metrics/internal/promdemo"
)

func main() {
	addr := ":2112"
	log.Printf("promdemo listening on %s", addr)
	if err := http.ListenAndServe(addr, promdemo.NewHandler()); err != nil {
		log.Fatal(err)
	}
}
