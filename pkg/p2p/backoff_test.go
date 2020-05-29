package p2p

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	b := defaultBackoff().(*backoff)
	// Initially the backoff should be min-backoff
	assert.Equal(t, b.min, b.Backoff())
	var i = 0
	for i = 0 ; i < 10; i++{
		b.Backoff().Milliseconds()
	}
	// In a few passes we should have reached max-backoff
	assert.Equal(t, b.max, b.Backoff())
}

func TestBackoffReset(t *testing.T) {
	b := defaultBackoff().(*backoff)
	b.Backoff()
	assert.True(t, b.Backoff() > b.min)
	b.Reset()
	assert.Equal(t, b.min, b.Backoff())
}

func TestBackoffDefaultValues(t *testing.T) {
	b := defaultBackoff().(*backoff)
	assert.Equal(t, time.Second, b.min)
	assert.Equal(t, 30 * time.Second, b.max)
}