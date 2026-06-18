package config

import (
	"os"
	"strconv"
)

type Config struct {
	AWSRegion            string
	SQSQueueURL          string
	DynamoTelemetryTable string
	DynamoAlertsTable    string

	Scenario           string
	SensorCount        int
	EventRatePerSecond int
	WindowSeconds      int

	RandomSeed         int64
	OutlierProbability float64
	MissingProbability float64

	EdgeGatewayHost string
	EdgeGatewayPort string
	CloudAPIPort    string

	WorkerBatchSize      int
	WorkerPollIntervalMs int
}

func Load() Config {
	return Config{
		AWSRegion:            getEnv("AWS_REGION", "us-east-1"),
		SQSQueueURL:          getEnv("SQS_QUEUE_URL", ""),
		DynamoTelemetryTable: getEnv("DYNAMODB_TELEMETRY_TABLE", "TelemetryAggregates"),
		DynamoAlertsTable:    getEnv("DYNAMODB_ALERTS_TABLE", "Alerts"),

		Scenario:           getEnv("SCENARIO", "edge_preprocessing"),
		SensorCount:        getEnvInt("SENSOR_COUNT", 10),
		EventRatePerSecond: getEnvInt("EVENT_RATE_PER_SECOND", 1),
		WindowSeconds:      getEnvInt("WINDOW_SECONDS", 5),

		RandomSeed:         getEnvInt64("RANDOM_SEED", 123456789),
		OutlierProbability: getEnvFloat("OUTLIER_PROBABILITY", 0.05),
		MissingProbability: getEnvFloat("MISSING_PROBABILITY", 0.02),

		EdgeGatewayHost: getEnv("EDGE_GATEWAY_HOST", "localhost"),
		EdgeGatewayPort: getEnv("EDGE_GATEWAY_PORT", "8081"),
		CloudAPIPort:    getEnv("CLOUD_API_PORT", "8080"),

		WorkerBatchSize:      getEnvInt("WORKER_BATCH_SIZE", 10),
		WorkerPollIntervalMs: getEnvInt("WORKER_POLL_INTERVAL_MS", 500),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
func getEnvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvFloat(key string, fallback float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}

	return parsed
}
