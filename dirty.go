// Package dirty provides clean unmarshalling of dirty JSON data.
// Dirty JSON data is JSON with unstable schema, flaky keys, etc.
package dirty

import (
	"bytes"
	"encoding/json"
)

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

func (e *Enabled) result() any { return e.res }
func (e *Enabled) init(v any)  { e.res = v }

// Disabled can mark your model as valid dirty.Model but won't enable dirtying.
// You can easily switch from `dirty.Enabled` to `dirty.Disabled` keeping all models & interfaces working
// but with pure json.Unmarshal only.
type Disabled struct{}

// Adjust Disabled logic, so actually we DO store the Green results
func (d *Disabled) result() any                    { return nil }
func (d *Disabled) init(v any)                     {}
func (d *Disabled) decode(dec *json.Decoder) error { return nil }

// isDisabled is check if we need to ignore dirty unmarshalling.
func (Disabled) isDisabled() { return }

// Unmarshal unmarshals given data via dirty Encoder
func Unmarshal(data []byte, v interface{}) error {
	r := bytes.NewReader(data)

	return NewDecoder(r).Decode(v)
}
