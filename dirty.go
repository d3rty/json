// Package dirty provides clean unmarshalling of dirty JSON data.
// Dirty JSON data is JSON with unstable schema, flaky keys, etc.
package dirty

import "bytes"

// Unmarshal parses the JSON-encoded data, allowing schema to be dirty,
// and stores the result in the value pointed to by v.
// It's the main part of public API, it's considered to be used instead of json.Unmarshal.
func Unmarshal(data []byte, v interface{}) error {
	r := bytes.NewReader(data)

	return NewDecoder(r).Decode(v)
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

func (e *Enabled) result() any { return e.res }
func (e *Enabled) init(v any)  { e.res = v }

// Disabled can mark your model as valid dirty.Model but won't enable dirtying.
// You can easily switch from `dirty.Enabled` to `dirty.Disabled` keeping all models & interfaces working
// but with pure json.Unmarshal only.
type Disabled struct{}

// Adjust Disabled logic, so actually we DO store the Green results
func (*Disabled) result() any { return nil }
func (*Disabled) init(_ any)  {}
func (*Disabled) isDisabled() {} // isDisabled disabled dirtying (keeping all interfaces working)
