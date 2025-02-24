package dirtyjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/d3rty/json/internal/config"
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

type Enabled struct {
	res any
}

func (e *Enabled) result() any { return e.res }
func (e *Enabled) init(v any)  { e.res = v }

// Disabled is an atom struct that that remains syntaxly valid dirty model,
// but disables dirty unmarshalling.
// You can easily switch from `dirty.Enabled` to `dirty.Disabled`
// keeping all models & interfaces working (falling back to standard (clean) json.Unmarshal).
type Disabled struct{}

func (*Disabled) result() any { return nil }
func (*Disabled) init(_ any)  {}
func (*Disabled) isDisabled() {} // isDisabled disabled dirtying (keeping all interfaces working)

type (
	// Number can marshall anything (that is possible) into float64.
	Number float64
	// String is just a string.
	String string
	// Bool can marshall anything (that is possible) into bool.
	Bool bool
	// Array for now is just for now is json's array.
	Array []any
	// Object for now is just for now is json's object.
	Object map[string]any

	// TODO: Arrays from String, Objects from strings. When some part of nested JSON is stringifed.
	// TODO: Time, Date, DateTime, etc.
	// TODO: Integer / Float
)

// UnmarshalJSON converts []byte into a Number.
func (v *Number) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("dirty.Number: UnmarshalJSON on nil pointer")
	}

	cfg := config.Get().Number

	var s string
	// If the value is a quoted string.
	if data[0] == '"' {
		if !cfg.AllowString {
			return fmt.Errorf("dirty.Number: string input not allowed")
		}
		if len(data) < 2 || data[len(data)-1] != '"' {
			return errors.New("dirty.Number: invalid string value")
		}
		s = string(data[1 : len(data)-1])
	} else {
		s = string(data)
	}

	// Remove spaces if allowed.
	if cfg.AllowSpacing {
		s = strings.ReplaceAll(s, " ", "")
	}
	// Remove commas if allowed.
	if cfg.AllowComma {
		s = strings.ReplaceAll(s, ",", "")
	}

	// Parse the float.
	n, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return fmt.Errorf("dirty.Number: cannot parse number: %w", err)
	}

	// we can't know about AllowFloatishIntegers here, as we don't know the destination clean type
	// (and we probably won't never know it here. so it will be at a later stage)

	*v = Number(n)
	return nil
}

// UnmarshalJSON converts []byte into a Bool.
func (v *String) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("dirty.String: UnmarshalJSON on nil pointer")
	}

	var s string
	if data[0] != '"' {
		return fmt.Errorf("dirty.String cant be parsed from: %s", s)
	}
	if len(data) < 2 {
		return errors.New("dirty.String missing closing quote")
	}
	if data[len(data)-1] != '"' {
		return errors.New("dirty.String missing closing quote")
	}

	s = string(data[1 : len(data)-1])
	*v = String(s)
	return nil

}

// UnmarshalJSON converts []byte into a Bool.
func (v *Bool) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("dirty.Bool: UnmarshalJSON on nil pointer")
	}

	cfg := config.Get().Bool

	var s string
	// Check if the incoming value is a quoted string.
	if data[0] == '"' {
		if !cfg.AllowString {
			return fmt.Errorf("dirty.Bool: string input not allowed")
		}
		if len(data) < 2 || data[len(data)-1] != '"' {
			return errors.New("dirty.Bool: invalid string value")
		}
		s = string(data[1 : len(data)-1])
	} else {
		// Otherwise, treat it as a raw token (number or boolean).
		s = string(data)
	}

	// Normalize the string.
	s = strings.TrimSpace(strings.ToLower(s))

	// If the input was a string (or we are allowing string interpretation)
	if cfg.AllowString {
		// Check against the configured true strings.
		for _, ts := range cfg.TrueStrings {
			if s == strings.ToLower(ts) {
				*v = true
				return nil
			}
		}
		// Check against the configured false strings.
		for _, fs := range cfg.FalseStrings {
			if s == strings.ToLower(fs) {
				*v = false
				return nil
			}
		}
		// If no match was found in string mode...
		if cfg.FallbackFromStringToRed {
			return fmt.Errorf("dirty.Bool: unrecognized string value %q", s)
		}
		*v = Bool(cfg.FallbackStringValue)
		return nil
	}

	// If string input is not allowed but number input is allowed.
	if cfg.AllowNumber {
		// Attempt to parse it as a float.
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("dirty.Bool: cannot parse numeric value %q: %w", s, err)
		}
		if cfg.TrueNumbers(n) {
			*v = true
			return nil
		}
		if cfg.FalseNumbers(n) {
			*v = false
			return nil
		}
		if cfg.FallbackFromNumberToRed {
			return fmt.Errorf("dirty.Bool: unrecognized numeric value %v", n)
		}
		*v = Bool(cfg.FallbackNumberValue)
		return nil
	}

	return fmt.Errorf("dirty.Bool: unsupported value %q", s)
}

// UnmarshalJSON converts []byte into an Array.
func (v *Array) UnmarshalJSON(data []byte) error {
	if v == nil {
		return nil
	}

	var s string
	if data[0] != '[' {
		if data[0] == 'n' { // null
			return nil
		}
		return fmt.Errorf("dirty.Array cant be parsed from: %s", s)
	}
	if len(data) < 2 {
		return errors.New("dirty.Array missing closing quote")
	}
	if data[len(data)-1] != ']' {
		return errors.New("dirty.Array missing closing quote")
	}

	var arr []any
	if err := json.Unmarshal(data[1:len(data)-1], &arr); err != nil {
		return fmt.Errorf("dirty.Array cant be parsed from json content: %w", err)
	}
	*v = Array(arr)
	return nil
}

// UnmarshalJSON converts []byte into an Object.
func (v *Object) UnmarshalJSON(data []byte) error {
	if v == nil {
		return nil
	}

	var s string
	if data[0] != '{' {
		if data[0] == 'n' { // null
			return nil
		}
		return fmt.Errorf("dirty.Object cant be parsed from: %s", s)
	}
	if len(data) < 2 {
		return errors.New("dirty.Object missing closing quote")
	}
	if data[len(data)-1] != '}' {
		return errors.New("dirty.Object missing closing quote")
	}

	var obj map[string]any
	if err := json.Unmarshal(data[1:len(data)-1], &obj); err != nil {
		return fmt.Errorf("dirty.Object cant be parsed from json content: %w", err)
	}
	*v = obj
	return nil
}
