package dirty_test

import (
	"testing"

	dirty "github.com/d3rty/json"
	"github.com/d3rty/json/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Event0 struct {
	ID       int  `json:"id"`
	IsActive bool `json:"is_active"`
}

type Event struct {
	dirty.Enabled // Step 1: Enabling dirty

	ID       int    `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`

	// MustBool won't be considered in dirty model, so it must parsed
	MustBool bool `json:"must_bool"`
}

type Envelope struct {
	Total  int     `json:"total"`
	Events []Event `json:"data"`
}

// Dirty links Event to EventDirty // Step 2: Linking dirty model.
func (e *Event) Dirty() any {
	return &EventDirty{}
}

type EventDirty struct {
	ID       dirty.Number `json:"id"`
	IsActive dirty.Bool   `json:"is_active"`
}

func TestUnmarshal_Green(t *testing.T) {
	var e Event0
	require.NoError(t, dirty.Unmarshal([]byte(`{"id":123, "is_active":true}`), &e))

	assert.Equal(t, 123, e.ID)
	assert.True(t, e.IsActive)
}

func TestUnmarshal_Yellow(t *testing.T) {
	var e Event
	require.NoError(t,
		dirty.Unmarshal([]byte(`{"id":"123","name":"foobar", "is_active":"on"}`), &e),
	)
	assert.Equal(t, 123, e.ID)
	assert.Equal(t, "foobar", e.Name)
	assert.True(t, e.IsActive)
}

func TestUnmarshal_Envelope(t *testing.T) {
	var e Envelope
	require.NoError(t,
		dirty.Unmarshal([]byte(`{"total":1,"data":[{"id":"123","name":"foobar","is_active":"1","must_bool":"true"}]}`), &e),
	)
	assert.Equal(t, 1, e.Total)
	assert.NotEmpty(t, e.Events)

	// It should be RED because of lost "must_bool" field

	evt := e.Events[0]
	assert.Equal(t, 123, evt.ID)
	assert.Equal(t, "foobar", evt.Name)
	assert.True(t, evt.IsActive)
	assert.False(t, evt.MustBool) // as it wasn't parsed as bool
}

func TestUnmarshal_EnvelopeFlexKeys(t *testing.T) {
	dirty.ConfigSetGlobal(func(cfg *dirty.Config) {
		cfg.ResetToEmpty()
		// only enable things we need here
		cfg.FlexKeys.Disabled = false
		cfg.FlexKeys.ChameleonCase = true
		cfg.FlexKeys.CaseInsensitive = true
		cfg.Number.Disabled = false
		cfg.Number.FromStrings.Disabled = false
		cfg.Bool.Disabled = false
		cfg.Bool.FromStrings.Disabled = false
		cfg.Bool.FromStrings.RespectFromNumbersLogic = true
		cfg.Bool.FromNumbers.Disabled = false
		cfg.Bool.FromNumbers.CustomParseFunc = config.BoolFromNumberBinary
	})

	var e Envelope
	require.NoError(t,
		dirty.Unmarshal([]byte(`{"total":1,"data":[{"id":"123","name":"foobar","Is-Active":"1","must_bool":"true"}]}`), &e),
	)
	assert.Equal(t, 1, e.Total)
	assert.NotEmpty(t, e.Events)

	// It should be RED because of lost "must_bool" field

	evt := e.Events[0]
	assert.Equal(t, 123, evt.ID)
	assert.Equal(t, "foobar", evt.Name)
	assert.True(t, evt.IsActive)
	assert.False(t, evt.MustBool) // as it wasn't parsed as bool
}
