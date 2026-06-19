package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"continuum-telemetry/internal/awsclient"
	"continuum-telemetry/internal/config"
	"continuum-telemetry/internal/edge"
	"continuum-telemetry/internal/messaging"
	"continuum-telemetry/internal/model"
)

var receivedEvents uint64
var validEvents uint64
var invalidEvents uint64
var localAnomalies uint64
var producedAggregates uint64

func main() {
	cfg := config.Load()

	windowDuration := time.Duration(cfg.WindowSeconds) * time.Second
	aggregator := edge.NewAggregator(windowDuration)
	thresholds := edge.DefaultAnomalyThresholds()
	publisher, err := buildPublisher(context.Background(), cfg)
	if err != nil {
		log.Fatalf("failed to initialize publisher: %v", err)
	}
	go flushExpiredWindowsLoop(aggregator, publisher)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /stats", statsHandler)
	mux.HandleFunc("POST /telemetry", func(w http.ResponseWriter, r *http.Request) {
		telemetryHandler(w, r, aggregator, thresholds, publisher)
	})

	addr := ":" + cfg.EdgeGatewayPort

	log.Println("edge-gateway started")
	log.Printf("listening_on=%s", addr)
	log.Printf("scenario=%s", cfg.Scenario)
	log.Printf("window_seconds=%d", cfg.WindowSeconds)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("edge-gateway stopped: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "edge-gateway",
	})
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]uint64{
		"received_events":     atomic.LoadUint64(&receivedEvents),
		"valid_events":        atomic.LoadUint64(&validEvents),
		"invalid_events":      atomic.LoadUint64(&invalidEvents),
		"local_anomalies":     atomic.LoadUint64(&localAnomalies),
		"produced_aggregates": atomic.LoadUint64(&producedAggregates),
	})
}

func telemetryHandler(
	w http.ResponseWriter,
	r *http.Request,
	aggregator *edge.Aggregator,
	thresholds edge.AnomalyThresholds,
	publisher messaging.Queue,
) {
	defer r.Body.Close()

	atomic.AddUint64(&receivedEvents, 1)

	var event model.TelemetryEvent

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		atomic.AddUint64(&invalidEvents, 1)

		writeJSON(w, http.StatusBadRequest, map[string]string{
			"status": "rejected",
			"error":  "invalid JSON payload",
		})
		return
	}

	if err := edge.ValidateTelemetryEvent(event); err != nil {
		atomic.AddUint64(&invalidEvents, 1)

		writeJSON(w, http.StatusBadRequest, map[string]string{
			"status": "rejected",
			"error":  err.Error(),
		})
		return
	}

	validCount := atomic.AddUint64(&validEvents, 1)

	isAnomaly := edge.IsLocalAnomaly(event, thresholds)
	if isAnomaly {
		atomic.AddUint64(&localAnomalies, 1)
	}

	aggregate, ready := aggregator.AddEvent(event, isAnomaly)
	if ready {
		publishAggregate(r.Context(), publisher, *aggregate)
	}

	if validCount%100 == 0 {
		log.Printf(
			"valid_events=%d latest_event_id=%s trace_id=%s sensor_id=%s temp=%.2f vibration=%.3f power=%.2f simulated_outlier=%v detected_anomaly=%v",
			validCount,
			event.EventID,
			event.TraceID,
			event.SensorID,
			event.Temperature,
			event.Vibration,
			event.Power,
			event.IsOutlier,
			isAnomaly,
		)
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":   "accepted",
		"event_id": event.EventID,
		"trace_id": event.TraceID,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}

func flushExpiredWindowsLoop(aggregator *edge.Aggregator, publisher messaging.Queue) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		aggregates := aggregator.FlushExpired(time.Now().UTC())

		for _, aggregate := range aggregates {
			publishAggregate(context.Background(), publisher, aggregate)
		}
	}
}
func publishAggregate(
	ctx context.Context,
	publisher messaging.Queue,
	aggregate model.TelemetryAggregate,
) {
	body, err := json.Marshal(aggregate)
	if err != nil {
		log.Printf("failed_to_marshal_aggregate aggregate_id=%s error=%v", aggregate.EventID, err)
		return
	}

	if err := publisher.Send(ctx, body); err != nil {
		log.Printf("failed_to_publish_aggregate aggregate_id=%s error=%v", aggregate.EventID, err)
		return
	}

	aggregateCount := atomic.AddUint64(&producedAggregates, 1)

	log.Printf(
		"aggregate_published=%d aggregate_id=%s sensor_id=%s events=%d avg_temp=%.2f max_temp=%.2f avg_vibration=%.3f avg_power=%.2f local_anomaly=%v",
		aggregateCount,
		aggregate.EventID,
		aggregate.SensorID,
		aggregate.EventCount,
		aggregate.AvgTemperature,
		aggregate.MaxTemperature,
		aggregate.AvgVibration,
		aggregate.AvgPower,
		aggregate.LocalAnomaly,
	)
}
func buildPublisher(ctx context.Context, cfg config.Config) (messaging.Queue, error) {
	switch cfg.MessagingBackend {
	case "log":
		log.Println("messaging_backend=log")
		return messaging.NewLogQueue("edge-aggregates"), nil

	case "sqs":
		if cfg.SQSQueueURL == "" {
			return nil, fmt.Errorf("SQS_QUEUE_URL is required when messaging backend is sqs")
		}

		awsCfg, err := awsclient.LoadConfig(ctx, cfg.AWSRegion, cfg.AWSEndpointURL)
		if err != nil {
			return nil, err
		}

		sqsClient := awsclient.NewSQSClient(awsCfg)

		log.Printf("messaging_backend=sqs queue_url=%s", cfg.SQSQueueURL)

		return messaging.NewSQSQueue(sqsClient, cfg.SQSQueueURL), nil

	default:
		return nil, fmt.Errorf("unknown messaging backend: %s", cfg.MessagingBackend)
	}
}
