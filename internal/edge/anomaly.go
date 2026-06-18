package edge

import "continuum-telemetry/internal/model"

type AnomalyThresholds struct {
	MaxTemperature float64
	MaxVibration   float64
	MaxPower       float64
}

func DefaultAnomalyThresholds() AnomalyThresholds {
	return AnomalyThresholds{
		MaxTemperature: 90.0,
		MaxVibration:   1.10,
		MaxPower:       600.0,
	}
}

func IsLocalAnomaly(event model.TelemetryEvent, thresholds AnomalyThresholds) bool {
	if event.Temperature >= thresholds.MaxTemperature {
		return true
	}

	if event.Vibration >= thresholds.MaxVibration {
		return true
	}

	if event.Power >= thresholds.MaxPower {
		return true
	}

	return false
}
