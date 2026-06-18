package edge

import (
	"fmt"
	"time"

	"continuum-telemetry/internal/model"
)

func ValidateTelemetryEvent(event model.TelemetryEvent) error {
	if event.EventID == "" {
		return fmt.Errorf("missing event_id")
	}

	if event.TraceID == "" {
		return fmt.Errorf("missing trace_id")
	}

	if event.SensorID == "" {
		return fmt.Errorf("missing sensor_id")
	}

	if event.Timestamp.IsZero() {
		return fmt.Errorf("missing timestamp")
	}

	if event.Scenario == "" {
		return fmt.Errorf("missing scenario")
	}

	if event.Timestamp.After(time.Now().UTC().Add(5 * time.Minute)) {
		return fmt.Errorf("timestamp is too far in the future")
	}

	if event.Temperature < -50 || event.Temperature > 200 {
		return fmt.Errorf("temperature out of physical range: %.2f", event.Temperature)
	}

	if event.Vibration < 0 || event.Vibration > 10 {
		return fmt.Errorf("vibration out of physical range: %.3f", event.Vibration)
	}

	if event.Power < 0 || event.Power > 5000 {
		return fmt.Errorf("power out of physical range: %.2f", event.Power)
	}

	return nil
}
