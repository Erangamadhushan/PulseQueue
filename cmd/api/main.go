package main

import (
	"log"
	"net/http"

	"github.com/Erangamadhushan/pulsequeue/internal/health"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", health.HealthHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("PulseQueue API starting on :8080")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}


}