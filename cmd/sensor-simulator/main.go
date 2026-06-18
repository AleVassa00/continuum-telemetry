package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"continuum-telemetry/internal/config"
	"continuum-telemetry/internal/model"
	"continuum-telemetry/internal/simulator"
)

func main() {
	cfg := config.Load()

	generator := simulator.NewGenerator(
		cfg.SensorCount,
		cfg.Scenario,
		cfg.EventRatePerSecond,
		cfg.RandomSeed,
		cfg.OutlierProbability,
		cfg.MissingProbability,
	)

	edgeURL := fmt.Sprintf(
		"http://%s:%s/telemetry",
		cfg.EdgeGatewayHost,
		cfg.EdgeGatewayPort,
	)

	log.Println("sensor-simulator started")
	log.Printf("scenario=%s", cfg.Scenario)
	log.Printf("sensor_count=%d", cfg.SensorCount)
	log.Printf("event_rate_per_second=%d", cfg.EventRatePerSecond)
	log.Printf("arrival_distribution=exponential")
	log.Printf("random_seed=%d", cfg.RandomSeed)
	log.Printf("outlier_probability=%.4f", cfg.OutlierProbability)
	log.Printf("missing_probability=%.4f", cfg.MissingProbability)
	log.Printf("edge_gateway_url=%s", edgeURL)

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	ctx := context.Background()

	var generatedEvents uint64
	var sentEvents uint64
	var failedEvents uint64
	var missingEvents uint64

	for {
		interArrival := generator.NextInterArrival()
		time.Sleep(interArrival)

		event, ok := generator.NextEvent()
		if !ok {
			missingEvents++

			if missingEvents%100 == 0 {
				log.Printf(
					"missing_events=%d generated_events=%d sent_events=%d failed_events=%d",
					missingEvents,
					generatedEvents,
					sentEvents,
					failedEvents,
				)
			}

			continue
		}

		generatedEvents++

		if err := sendTelemetryEvent(ctx, httpClient, edgeURL, event); err != nil {
			failedEvents++

			log.Printf(
				"failed to send event_id=%s trace_id=%s sensor_id=%s error=%v",
				event.EventID,
				event.TraceID,
				event.SensorID,
				err,
			)

			continue
		}

		sentEvents++

		if sentEvents%100 == 0 {
			log.Printf(
				"generated=%d sent=%d failed=%d missing=%d latest_sensor=%s inter_arrival=%s temp=%.2f vibration=%.3f power=%.2f outlier=%v",
				generatedEvents,
				sentEvents,
				failedEvents,
				missingEvents,
				event.SensorID,
				interArrival,
				event.Temperature,
				event.Vibration,
				event.Power,
				event.IsOutlier,
			)
		}
	}
}

func sendTelemetryEvent(
	ctx context.Context,
	client *http.Client,
	url string,
	event model.TelemetryEvent,
) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return nil
}
