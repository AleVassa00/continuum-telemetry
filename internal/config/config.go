package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	AWSRegion            string
	AWSEndpointURL       string
	SQSQueueURL          string
	DynamoTelemetryTable string
	DynamoAlertsTable    string

	Scenario            string
	SensorCount         int
	EventRatePerSecond  int
	ArrivalDistribution string
	WindowSeconds       int

	RandomSeed         int64
	OutlierProbability float64
	MissingProbability float64

	EdgeGatewayHost string
	EdgeGatewayPort string
	CloudAPIPort    string

	MessagingBackend string

	WorkerBatchSize      int
	WorkerPollIntervalMs int
}

type yamlConfig struct {
	Scenario string `yaml:"scenario"`

	SensorSimulator struct {
		SensorCount         *int     `yaml:"sensor_count"`
		EventRatePerSecond  *int     `yaml:"event_rate_per_second"`
		ArrivalDistribution string   `yaml:"arrival_distribution"`
		RandomSeed          *int64   `yaml:"random_seed"`
		OutlierProbability  *float64 `yaml:"outlier_probability"`
		MissingProbability  *float64 `yaml:"missing_probability"`
	} `yaml:"sensor_simulator"`

	EdgeGateway struct {
		Host          string `yaml:"host"`
		Port          *int   `yaml:"port"`
		WindowSeconds *int   `yaml:"window_seconds"`
	} `yaml:"edge_gateway"`

	CloudAPI struct {
		Port *int `yaml:"port"`
	} `yaml:"cloud_api"`

	Worker struct {
		BatchSize      *int `yaml:"batch_size"`
		PollIntervalMs *int `yaml:"poll_interval_ms"`
	} `yaml:"worker"`

	Messaging struct {
		Backend string `yaml:"backend"`
	} `yaml:"messaging"`
}

func Load() Config {
	_ = godotenv.Load()

	cfg := defaultConfig()

	configPath := getEnv("APP_CONFIG", "config/local.yml")
	if err := applyYAMLConfig(&cfg, configPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load config file %s: %v\n", configPath, err)
	}

	applyEnvOverrides(&cfg)

	return cfg
}

func defaultConfig() Config {
	return Config{
		AWSRegion:            "us-east-1",
		AWSEndpointURL:       "",
		SQSQueueURL:          "",
		DynamoTelemetryTable: "TelemetryAggregates",
		DynamoAlertsTable:    "Alerts",

		Scenario:            "edge_preprocessing",
		SensorCount:         10,
		EventRatePerSecond:  1,
		ArrivalDistribution: "exponential",
		WindowSeconds:       5,

		RandomSeed:         123456789,
		OutlierProbability: 0.05,
		MissingProbability: 0.02,

		EdgeGatewayHost: "localhost",
		EdgeGatewayPort: "8081",
		CloudAPIPort:    "8080",

		WorkerBatchSize:      10,
		WorkerPollIntervalMs: 500,
		MessagingBackend:     "log",
	}
}

func applyYAMLConfig(cfg *Config, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var yc yamlConfig
	if err := yaml.Unmarshal(content, &yc); err != nil {
		return err
	}

	if yc.Scenario != "" {
		cfg.Scenario = yc.Scenario
	}

	if yc.SensorSimulator.SensorCount != nil {
		cfg.SensorCount = *yc.SensorSimulator.SensorCount
	}

	if yc.SensorSimulator.EventRatePerSecond != nil {
		cfg.EventRatePerSecond = *yc.SensorSimulator.EventRatePerSecond
	}

	if yc.SensorSimulator.ArrivalDistribution != "" {
		cfg.ArrivalDistribution = yc.SensorSimulator.ArrivalDistribution
	}

	if yc.SensorSimulator.RandomSeed != nil {
		cfg.RandomSeed = *yc.SensorSimulator.RandomSeed
	}

	if yc.SensorSimulator.OutlierProbability != nil {
		cfg.OutlierProbability = *yc.SensorSimulator.OutlierProbability
	}

	if yc.SensorSimulator.MissingProbability != nil {
		cfg.MissingProbability = *yc.SensorSimulator.MissingProbability
	}

	if yc.EdgeGateway.Host != "" {
		cfg.EdgeGatewayHost = yc.EdgeGateway.Host
	}

	if yc.EdgeGateway.Port != nil {
		cfg.EdgeGatewayPort = strconv.Itoa(*yc.EdgeGateway.Port)
	}

	if yc.EdgeGateway.WindowSeconds != nil {
		cfg.WindowSeconds = *yc.EdgeGateway.WindowSeconds
	}

	if yc.CloudAPI.Port != nil {
		cfg.CloudAPIPort = strconv.Itoa(*yc.CloudAPI.Port)
	}

	if yc.Worker.BatchSize != nil {
		cfg.WorkerBatchSize = *yc.Worker.BatchSize
	}

	if yc.Worker.PollIntervalMs != nil {
		cfg.WorkerPollIntervalMs = *yc.Worker.PollIntervalMs
	}
	if yc.Messaging.Backend != "" {
		cfg.MessagingBackend = yc.Messaging.Backend
	}

	return nil
}

func applyEnvOverrides(cfg *Config) {
	cfg.AWSRegion = getEnv("AWS_REGION", cfg.AWSRegion)
	cfg.AWSEndpointURL = getEnv("AWS_ENDPOINT_URL", cfg.AWSEndpointURL)
	cfg.SQSQueueURL = getEnv("SQS_QUEUE_URL", cfg.SQSQueueURL)
	cfg.DynamoTelemetryTable = getEnv("DYNAMODB_TELEMETRY_TABLE", cfg.DynamoTelemetryTable)
	cfg.DynamoAlertsTable = getEnv("DYNAMODB_ALERTS_TABLE", cfg.DynamoAlertsTable)
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
