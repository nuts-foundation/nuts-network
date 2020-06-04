package api

import (
	"github.com/magiconair/properties/assert"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-network/test"
	assert2 "github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDocument_fromModel(t *testing.T) {
	expected := model.Document{
		Hash:      test.StringToHash("37076f2cbe014a109d79b61ae9c7191f4cc57afc"),
		Type:      "test",
		Timestamp: time.Now(),
	}
	actual := Document{}
	actual.fromModel(expected)
	assert.Equal(t, actual.Type, expected.Type)
	assert.Equal(t, actual.Hash, expected.Hash.String())
	assert.Equal(t, actual.Timestamp, expected.Timestamp.UnixNano())
}

func TestDocument_toModel(t *testing.T) {
	expected := model.Document{
		Hash:      test.StringToHash("37076f2cbe014a109d79b61ae9c7191f4cc57afc"),
		Type:      "test",
		Timestamp: time.Unix(0, 1591255458635412720),
	}
	actual := Document{
		Hash:      "37076f2cbe014a109d79b61ae9c7191f4cc57afc",
		Timestamp: 1591255458635412720,
		Type:      "test",
	}
	assert2.Equal(t, expected, actual.toModel())
}
