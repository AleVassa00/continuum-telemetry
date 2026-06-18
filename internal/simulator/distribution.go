package simulator

import "time"

const (
	streamArrival = iota
	streamSensorSelection
	streamMissing
	streamOutlier
	streamTemperature
	streamVibration
	streamPower
)

type Distribution struct {
	rngs *Rngs
	rvgs *Variates
}

func NewDistribution(seed int64) *Distribution {
	rngs := NewRngs(seed)

	return &Distribution{
		rngs: rngs,
		rvgs: NewVariates(rngs),
	}
}

func (d *Distribution) ExponentialInterArrival(meanSeconds float64) time.Duration {
	d.rngs.SelectStream(streamArrival)

	seconds := d.rvgs.Exponential(meanSeconds)

	return time.Duration(seconds * float64(time.Second))
}

func (d *Distribution) SensorIndex(sensorCount int) int {
	d.rngs.SelectStream(streamSensorSelection)

	return d.rvgs.Equilikely(1, sensorCount)
}

func (d *Distribution) Missing(probability float64) bool {
	d.rngs.SelectStream(streamMissing)

	return d.rvgs.Bernoulli(probability)
}

func (d *Distribution) Outlier(probability float64) bool {
	d.rngs.SelectStream(streamOutlier)

	return d.rvgs.Bernoulli(probability)
}

func (d *Distribution) Temperature(isOutlier bool) float64 {
	d.rngs.SelectStream(streamTemperature)

	if isOutlier {
		return d.rvgs.Normal(105.0, 5.0)
	}

	return d.rvgs.Normal(72.0, 3.0)
}

func (d *Distribution) Vibration(isOutlier bool) float64 {
	d.rngs.SelectStream(streamVibration)

	var value float64

	if isOutlier {
		value = d.rvgs.Normal(1.45, 0.20)
	} else {
		value = d.rvgs.Normal(0.60, 0.08)
	}

	if value < 0 {
		return 0
	}

	return value
}

func (d *Distribution) Power(isOutlier bool) float64 {
	d.rngs.SelectStream(streamPower)

	if isOutlier {
		return d.rvgs.Normal(700.0, 80.0)
	}

	return d.rvgs.Normal(400.0, 25.0)
}
