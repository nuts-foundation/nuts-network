package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMarshalDocumentTime(t *testing.T) {
	assert.Equal(t, int64(0), MarshalDocumentTime(time.Time{}))
	assert.Equal(t, int64(1603225800000000000), MarshalDocumentTime(time.Date(2020, 10, 20, 20, 30, 0, 0, time.UTC)))
}

func TestUnmarshalDocumentTime(t *testing.T) {
	assert.True(t, UnmarshalDocumentTime(0).IsZero())
	assert.Equal(t, time.Date(2020, 10, 20, 20, 30, 0, 0, time.UTC), UnmarshalDocumentTime(1603225800000000000))
}
