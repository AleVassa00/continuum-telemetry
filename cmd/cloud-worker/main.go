package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"continuum-telemetry/internal/awsclient"
	"continuum-telemetry/internal/config"
	"continuum-telemetry/internal/messaging"
	"continuum-telemetry/internal/storage"
	"continuum-telemetry/internal/worker"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if cfg.MessagingBackend != "sqs" {
		log.Fatalf("cloud-worker requires messaging backend sqs, got %s", cfg.MessagingBackend)
	}

	if cfg.SQSQueueURL == "" {
		log.Fatal("SQS_QUEUE_URL is required")
	}

	awsCfg, err := awsclient.LoadConfig(ctx, cfg.AWSRegion, cfg.AWSEndpointURL)
	if err != nil {
		log.Fatalf("failed to load aws config: %v", err)
	}

	sqsClient := awsclient.NewSQSClient(awsCfg)
	dynamoClient := awsclient.NewDynamoDBClient(awsCfg)

	queue := messaging.NewSQSQueue(sqsClient, cfg.SQSQueueURL)
	store := storage.NewDynamoStore(dynamoClient, cfg.DynamoTelemetryTable)

	processor := worker.NewProcessor(
		queue,
		store,
		cfg.WorkerBatchSize,
		time.Duration(cfg.WorkerPollIntervalMs)*time.Millisecond,
	)

	log.Printf("aws_region=%s aws_endpoint=%s", cfg.AWSRegion, cfg.AWSEndpointURL)
	log.Printf("sqs_queue_url=%s", cfg.SQSQueueURL)
	log.Printf("dynamodb_table=%s", cfg.DynamoTelemetryTable)

	if err := processor.Run(ctx); err != nil && ctx.Err() == nil {
		log.Fatalf("cloud-worker stopped with error: %v", err)
	}
}
