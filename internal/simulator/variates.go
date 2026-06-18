package simulator

import "math"

type Variates struct {
	rngs *Rngs
}

func NewVariates(rngs *Rngs) *Variates {
	return &Variates{
		rngs: rngs,
	}
}

func (v *Variates) Bernoulli(p float64) bool {
	if p <= 0 {
		return false
	}

	if p >= 1 {
		return true
	}

	return v.rngs.Random() < p
}

func (v *Variates) Equilikely(a int, b int) int {
	if b <= a {
		return a
	}

	u := v.rngs.Random()

	return a + int(float64(b-a+1)*u)
}

func (v *Variates) Uniform(a float64, b float64) float64 {
	if b <= a {
		return a
	}

	return a + (b-a)*v.rngs.Random()
}

func (v *Variates) Exponential(mean float64) float64 {
	if mean <= 0 {
		return 0
	}

	u := v.rngs.Random()

	return -mean * math.Log(1.0-u)
}

func (v *Variates) Normal(mean float64, stddev float64) float64 {
	if stddev <= 0 {
		return mean
	}

	u1 := v.rngs.Random()
	u2 := v.rngs.Random()

	z := math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)

	return mean + stddev*z
}
