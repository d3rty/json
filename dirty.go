// Package dirty provides clean unmarshalling of dirty JSON data.
// Dirty JSON data is JSON with unstable schema, flaky keys, etc.
package dirty

import (
	"bytes"

	"github.com/d3rty/json/internal/config"
	"github.com/d3rty/json/internal/dirtyjson"
)

// Unmarshal parses the JSON-encoded data, allowing schema to be dirty,
// and stores the result in the value pointed to by v.
// It's the main part of public API, it's considered to be used instead of json.Unmarshal.
func Unmarshal(data []byte, v any) error {
	r := bytes.NewReader(data)

	return dirtyjson.NewDecoder(r).Decode(v)
}

//
// Config
//

//nolint:gochecknoglobals // we're fine with these global aliases
var (
	// ConfigSetGlobal allows us to update the global config.
	ConfigSetGlobal = config.SetGlobal

	// ConfigFromBytes parses a TOML configuration from bytes.
	ConfigFromBytes = config.FromBytes
)

type Config = config.Config

// Number is a custom type for unmarshalling numbers.
// Numbers can be parsed from strings or actual JSON numbers.
// Other types will be rejected.
type Number = dirtyjson.Number

// String s a custom type for unmarshalling strings.
// Anything except an actual JSON string will be rejected.
type String = dirtyjson.String

// Bool is a custom type for unmarshalling booleans.
// Bools can be parsed from
//   - strings ("true", "false", "yes", "no", "on", "off", "1", "0")
//   - numbers (1, 0)
//   - actual JSON booleans.
//
// Other types will be rejected.
type Bool = dirtyjson.Bool

// Array is a custom type for unmarshalling arrays.
// Anything except actual JSON arrays will be rejected.
type Array = dirtyjson.Array

// Object is a custom type for unmarshalling objects.
// Anything except actual JSON objects will be rejected.
type Object = dirtyjson.Object

// Enabled is an atom struct that enables dirty unmarshalling
// for a given clean model. It MUST be embedded into any clean model.
// The clean model also MUST implement Dirtyable interface.
type Enabled = dirtyjson.Enabled

// Disabled is an atom struct that that remains a syntactically valid dirty model
// but disables dirty unmarshalling.
// You can easily switch from `dirty.Enabled` to `dirty.Disabled`
// keeping all models and interfaces working (falling back to standard (clean) json.Unmarshal).
type Disabled = dirtyjson.Disabled
