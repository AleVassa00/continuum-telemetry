package simulator

import "time"

const (
	modulus     int64 = 2147483647
	multiplier  int64 = 48271
	defaultSeed int64 = 123456789
	streams           = 256
	a256        int64 = 22925
)

type Rngs struct {
	seed        [streams]int64
	stream      int
	initialized bool
}

func NewRngs(seed int64) *Rngs {
	r := &Rngs{}
	r.seed[0] = defaultSeed

	if seed != 0 {
		r.PlantSeeds(seed)
	}

	return r
}

func (r *Rngs) Random() float64 {
	q := modulus / multiplier
	rem := modulus % multiplier

	t := multiplier*(r.seed[r.stream]%q) - rem*(r.seed[r.stream]/q)

	if t > 0 {
		r.seed[r.stream] = t
	} else {
		r.seed[r.stream] = t + modulus
	}

	return float64(r.seed[r.stream]) / float64(modulus)
}

func (r *Rngs) SelectStream(index int) {
	if index < 0 {
		index = -index
	}

	r.stream = index % streams

	if !r.initialized && r.stream != 0 {
		r.PlantSeeds(defaultSeed)
	}
}

func (r *Rngs) PlantSeeds(x int64) {
	r.initialized = true

	currentStream := r.stream

	r.stream = 0
	r.PutSeed(x)

	r.stream = currentStream

	q := modulus / a256
	rem := modulus % a256

	for j := 1; j < streams; j++ {
		next := a256*(r.seed[j-1]%q) - rem*(r.seed[j-1]/q)

		if next > 0 {
			r.seed[j] = next
		} else {
			r.seed[j] = next + modulus
		}
	}
}

func (r *Rngs) PutSeed(x int64) {
	if x > 0 {
		x = x % modulus
	}

	if x < 0 {
		x = time.Now().UnixNano() % modulus
	}

	if x == 0 {
		x = defaultSeed
	}

	r.seed[r.stream] = x
}

func (r *Rngs) GetSeed() int64 {
	return r.seed[r.stream]
}
