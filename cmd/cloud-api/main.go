package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"continuum-telemetry/internal/api"
	"continuum-telemetry/internal/awsclient"
	"continuum-telemetry/internal/config"
	"continuum-telemetry/internal/storage"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	awsCfg, err := awsclient.LoadConfig(ctx, cfg.AWSRegion, cfg.AWSEndpointURL)
	if err != nil {
		log.Fatalf("failed to load aws config: %v", err)
	}

	dynamoClient := awsclient.NewDynamoDBClient(awsCfg)
	store := storage.NewDynamoStore(dynamoClient, cfg.DynamoTelemetryTable)

	apiServer := api.NewServer(store)

	httpServer := &http.Server{
		Addr:              ":" + cfg.CloudAPIPort,
		Handler:           apiServer.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("cloud-api started")
	log.Printf("listening_on=:%s", cfg.CloudAPIPort)
	log.Printf("aws_region=%s aws_endpoint=%s", cfg.AWSRegion, cfg.AWSEndpointURL)
	log.Printf("dynamodb_table=%s", cfg.DynamoTelemetryTable)

	errCh := make(chan error, 1)

	go func() {
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("cloud-api shutdown error: %v", err)
		}

	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("cloud-api stopped with error: %v", err)
		}
	}
}
