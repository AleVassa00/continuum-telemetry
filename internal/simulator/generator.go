package simulator

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"continuum-telemetry/internal/model"
)

type Generator struct {
	sensorCount             int
	scenario                string
	meanInterArrivalSeconds float64
	outlierProbability      float64
	missingProbability      float64
	distribution            *Distribution
}

func NewGenerator(
	sensorCount int,
	scenario string,
	eventRatePerSecond int,
	seed int64,
	outlierProbability float64,
	missingProbability float64,
) *Generator {
	if sensorCount <= 0 {
		sensorCount = 1
	}

	if eventRatePerSecond <= 0 {
		eventRatePerSecond = 1
	}

	meanInterArrivalSeconds := 1.0 / float64(eventRatePerSecond)

	return &Generator{
		sensorCount:             sensorCount,
		scenario:                scenario,
		meanInterArrivalSeconds: meanInterArrivalSeconds,
		outlierProbability:      outlierProbability,
		missingProbability:      missingProbability,
		distribution:            NewDistribution(seed),
	}
}

func (g *Generator) NextInterArrival() time.Duration {
	return g.distribution.ExponentialInterArrival(g.meanInterArrivalSeconds)
}

func (g *Generator) NextEvent() (model.TelemetryEvent, bool) {
	if g.distribution.Missing(g.missingProbability) {
		return model.TelemetryEvent{}, false
	}

	sensorIndex := g.distribution.SensorIndex(g.sensorCount)
	sensorID := fmt.Sprintf("machine-%03d", sensorIndex)

	isOutlier := g.distribution.Outlier(g.outlierProbability)

	event := model.TelemetryEvent{
		EventID:     newID(),
		TraceID:     newID(),
		SensorID:    sensorID,
		Timestamp:   time.Now().UTC(),
		Scenario:    g.scenario,
		Temperature: g.distribution.Temperature(isOutlier),
		Vibration:   g.distribution.Vibration(isOutlier),
		Power:       g.distribution.Power(isOutlier),
		IsOutlier:   isOutlier,
	}

	return event, true
}

func newID() string {
	bytes := make([]byte, 16)

	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return hex.EncodeToString(bytes)
}
