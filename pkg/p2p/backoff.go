package p2p

import (
	"time"
)

type Backoff interface {
	Reset()
	Backoff() time.Duration
}

type backoff struct {
	multiplier float64
	value      time.Duration
	max        time.Duration
	min        time.Duration
}

func (b *backoff) Reset() {
	b.value = 0
}

func (b *backoff) Backoff() time.Duration {
	// TODO: Might want to add jitter (e.g. https://github.com/grpc/grpc/blob/master/doc/connection-backoff.md)
	if b.value < b.min {
		b.value = b.min
	} else {
		b.value = time.Duration(float64(b.value) * b.multiplier)
		if b.value > b.max {
			b.value = b.max
		}
	}
	return b.value
}

func defaultBackoff() Backoff {
	// TODO: Make this configurable
	return &backoff{
		multiplier: 1.5,
		value:      0,
		max:        30 * time.Second,
		min:        time.Second,
	}
}
