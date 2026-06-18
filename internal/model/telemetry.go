package model

import "time"

type TelemetryEvent struct {
	EventID   string    `json:"event_id"`
	TraceID   string    `json:"trace_id"`
	SensorID  string    `json:"sensor_id"`
	Timestamp time.Time `json:"timestamp"`
	Scenario  string    `json:"scenario"`

	Temperature float64 `json:"temperature"`
	Vibration   float64 `json:"vibration"`
	Power       float64 `json:"power"`

	IsOutlier bool `json:"is_outlier"`
}

type TelemetryAggregate struct {
	EventID     string    `json:"event_id"`
	TraceID     string    `json:"trace_id"`
	SensorID    string    `json:"sensor_id"`
	WindowStart time.Time `json:"window_start"`
	WindowEnd   time.Time `json:"window_end"`
	Scenario    string    `json:"scenario"`

	AvgTemperature float64 `json:"avg_temperature"`
	MaxTemperature float64 `json:"max_temperature"`
	AvgVibration   float64 `json:"avg_vibration"`
	AvgPower       float64 `json:"avg_power"`

	EventCount   int  `json:"event_count"`
	LocalAnomaly bool `json:"local_anomaly"`
}

type Alert struct {
	AlertID   string    `json:"alert_id"`
	EventID   string    `json:"event_id"`
	TraceID   string    `json:"trace_id"`
	SensorID  string    `json:"sensor_id"`
	Timestamp time.Time `json:"timestamp"`
	Severity  string    `json:"severity"`
	Reason    string    `json:"reason"`
}
