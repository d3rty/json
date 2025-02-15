// Package dirty provides clean unmarshalling of dirty JSON data.
// Dirty JSON data is JSON with unstable schema, flaky keys, etc.
package dirty

import (
	"encoding/json"
	"errors"
)

// d3rtyMarker is an internal marker interface
// it allows Unmarshal function to validate that it's a custom dirty model.
type d3rtyMarker interface {
	unmarshal([]byte) error
}

// d3rtyContainer is an internal interface
// that allow us to init and retrieve dirty data
type d3rtyContainer interface {
	init(any)
	result() any
}

// Dirtyable is a clean model that has dirty model attached
// It's used as a way to link clean model with dirty model
type Dirtyable interface {
	Dirty() any
}

// Enabled is an atom struct that enables dirty unmarshalling for a given clean model.
// It MUST be embedded into any clean model (clean model also MUST implement Dirtyable interface)
type Enabled struct {
	res any
}

var _ d3rtyMarker = (*Enabled)(nil)

func (e *Enabled) result() any                 { return e.res }
func (e *Enabled) init(v any)                  { e.res = v }
func (e *Enabled) unmarshal(data []byte) error { return json.Unmarshal(data, e.res) }

// Disabled can mark your model as valid dirty.Model but won't enable dirtying.
// You can easily switch from `dirty.Enabled` to `dirty.Disabled` keeping all models & interfaces working
// but with pure json.Unmarshal only.
type Disabled struct{}

// Adjust Disabled logic, so actually we DO store the Green results
func (d *Disabled) result() any                 { return nil }
func (d *Disabled) init(v any)                  {}
func (d *Disabled) unmarshal(data []byte) error { return nil }

// isDisabled is check if we need to ignore dirty unmarshalling.
func (Disabled) isDisabled() { return }

// Unmarshal is the main Unmarshal function
func Unmarshal(data []byte, clean any) error {
	// Green phase: we could convert directly to clean
	var err error
	if err = json.Unmarshal(data, clean); err == nil {
		return nil
	}

	// Before starting dirty unmarshalling let's ensure user manually didn't disable dirtying
	// If yes - simply return original json error.
	if _, ok := clean.(interface{ isDisabled() }); ok {
		return err
	}

	// Let's ensure user configured everything correctly for dirty unmarshalling
	// If not - that was a regular unmarshalling, simply return original json error.

	schemer, ok := clean.(Dirtyable)
	if !ok {
		return err
	}
	container, ok := clean.(d3rtyContainer)
	if !ok {
		return err
	}

	// Yellow phase: try unmarshal into dirty model

	scheme := schemer.Dirty()
	container.init(scheme)

	if err := container.(d3rtyMarker).unmarshal(data); err != nil {
		// RED Phase: we couldn't unmarshal even in dirty model
		// TODO: we should Unmarshal in map or []map and rebuild dirty model
		return errors.New("red:to be implemented 2")
	}

	// Here comes the slow part, refactor it so it's fast
	// Currently for idea purpose that's OK
	buffer, err := json.Marshal(scheme)
	if err != nil {
		return errors.New("fixme 1")
	}
	if err := json.Unmarshal(buffer, clean); err != nil {
		return errors.New("fixme 2")
	}

	// Yellow Phase: OK: converting from dirty into clean model
	return nil
}
