package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Toskies/go-metrics/internal/otelstackdemo"
)

func main() {
	ctx := context.Background()

	cfg := otelstackdemo.DefaultConfig()
	providers, err := otelstackdemo.Setup(ctx, cfg)
	if err != nil {
		log.Fatalf("setup telemetry: %v", err)
	}

	obs, err := otelstackdemo.NewObservability(
		providers.Tracer("github.com/Toskies/go-metrics/internal/otelstackdemo"),
		providers.Meter("github.com/Toskies/go-metrics/internal/otelstackdemo"),
	)
	if err != nil {
		log.Fatalf("build observability: %v", err)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: otelstackdemo.NewHandler(obs),
	}

	go func() {
		log.Printf("otelstackdemo listening on %s", srv.Addr)
		log.Printf("OTLP endpoint: %s", cfg.OTLPEndpoint)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}
	if err := providers.Shutdown(shutdownCtx); err != nil {
		log.Printf("telemetry shutdown error: %v", err)
	}
}
