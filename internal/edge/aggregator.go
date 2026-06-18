package edge

import (
	"fmt"
	"sync"
	"time"

	"continuum-telemetry/internal/model"
)

type Aggregator struct {
	windowDuration time.Duration
	mu             sync.Mutex
	windows        map[string]*sensorWindow
}

type sensorWindow struct {
	sensorID string

	windowStart time.Time
	windowEnd   time.Time

	traceID  string
	scenario string

	sumTemperature float64
	maxTemperature float64
	sumVibration   float64
	sumPower       float64

	eventCount   int
	localAnomaly bool
}

func NewAggregator(windowDuration time.Duration) *Aggregator {
	if windowDuration <= 0 {
		windowDuration = 5 * time.Second
	}

	return &Aggregator{
		windowDuration: windowDuration,
		windows:        make(map[string]*sensorWindow),
	}
}

func (a *Aggregator) AddEvent(event model.TelemetryEvent, localAnomaly bool) (*model.TelemetryAggregate, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	currentWindow, exists := a.windows[event.SensorID]
	if !exists {
		a.windows[event.SensorID] = newSensorWindow(event, a.windowDuration, localAnomaly)
		return nil, false
	}

	if !event.Timestamp.Before(currentWindow.windowEnd) {
		aggregate := currentWindow.toAggregate()

		a.windows[event.SensorID] = newSensorWindow(event, a.windowDuration, localAnomaly)

		return &aggregate, true
	}

	currentWindow.add(event, localAnomaly)

	return nil, false
}

func (a *Aggregator) FlushExpired(now time.Time) []model.TelemetryAggregate {
	a.mu.Lock()
	defer a.mu.Unlock()

	aggregates := make([]model.TelemetryAggregate, 0)

	for sensorID, currentWindow := range a.windows {
		if now.Before(currentWindow.windowEnd) {
			continue
		}

		aggregate := currentWindow.toAggregate()
		aggregates = append(aggregates, aggregate)

		delete(a.windows, sensorID)
	}

	return aggregates
}

func newSensorWindow(event model.TelemetryEvent, duration time.Duration, localAnomaly bool) *sensorWindow {
	windowStart := event.Timestamp.Truncate(duration)

	return &sensorWindow{
		sensorID: event.SensorID,

		windowStart: windowStart,
		windowEnd:   windowStart.Add(duration),

		traceID:  event.TraceID,
		scenario: event.Scenario,

		sumTemperature: event.Temperature,
		maxTemperature: event.Temperature,
		sumVibration:   event.Vibration,
		sumPower:       event.Power,

		eventCount:   1,
		localAnomaly: localAnomaly,
	}
}

func (w *sensorWindow) add(event model.TelemetryEvent, localAnomaly bool) {
	w.sumTemperature += event.Temperature
	w.sumVibration += event.Vibration
	w.sumPower += event.Power
	w.eventCount++

	if event.Temperature > w.maxTemperature {
		w.maxTemperature = event.Temperature
	}

	if localAnomaly {
		w.localAnomaly = true
	}
}

func (w *sensorWindow) toAggregate() model.TelemetryAggregate {
	return model.TelemetryAggregate{
		EventID: fmt.Sprintf(
			"agg-%s-%d",
			w.sensorID,
			w.windowStart.UnixNano(),
		),
		TraceID:     w.traceID,
		SensorID:    w.sensorID,
		WindowStart: w.windowStart,
		WindowEnd:   w.windowEnd,
		Scenario:    w.scenario,

		AvgTemperature: w.sumTemperature / float64(w.eventCount),
		MaxTemperature: w.maxTemperature,
		AvgVibration:   w.sumVibration / float64(w.eventCount),
		AvgPower:       w.sumPower / float64(w.eventCount),

		EventCount:   w.eventCount,
		LocalAnomaly: w.localAnomaly,
	}
}
