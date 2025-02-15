package dirty_test

import (
	"testing"

	dirty "github.com/d3rty/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Event0 struct {
	ID int `json:"id"`
}

type Event struct {
	dirty.Enabled // Step 1: Enabling dirty

	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Dirty links Event to EventDirty // Step 2: Linking dirty model
func (e *Event) Dirty() any {
	return &EventDirty{}
}

type EventDirty struct {
	ID dirty.Number `json:"id"`
}

func ExampleUnmarshal() {
	// Step 3: Safe dirty unmarshal. "123" will be parsed in clean model as 123
	var e Event
	err := dirty.Unmarshal([]byte(`{"id":"123"}`), &e)

	_ = err // err happens when couldn't do anything

	result := dirty.ExtractResult[EventDirty](&e)

	_ = result.Color() // Here will be YELLOW
	// Green - parsed 100% directly to clean
	// Yellow - parsed without loss to dirty model
	// Red - partially parsed (with losses) into dirty model

	result.Warnings() // in case of yellow: warnings e.g. "123" -> 123
	result.Errors()   // in case of red: losses of data (fields unrecognized, unknown types, etc)

}

func TestUnmarshal_Green(t *testing.T) {
	var e Event0
	require.NoError(t, dirty.Unmarshal([]byte(`{"id":123}`), &e))

	assert.Equal(t, 123, e.ID)
}

func TestUnmarshal_Yellow(t *testing.T) {
	var e Event
	require.NoError(t,
		dirty.Unmarshal([]byte(`{"id":"123","name":"foobar"}`), &e),
	)
	assert.Equal(t, 123, e.ID)
	assert.Equal(t, "foobar", e.Name)

	result := dirty.ExtractResult[EventDirty](&e)

	assert.Equal(t, dirty.ColorYellow, result.Color())

	assert.Empty(t, result.Warnings()) // TODO warnings must be 1
	assert.Empty(t, result.Errors())
}
