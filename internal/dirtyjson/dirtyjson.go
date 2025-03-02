package dirtyjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/d3rty/json/internal/config"
	"github.com/d3rty/json/internal/option"
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

	cfg := config.Global().Number
	if !cfg.Allowed { // TODO: that's bad. do it better
		cfg.FromNull.Allowed = false
		cfg.FromStrings.Allowed = false
		cfg.FromNull.Allowed = false
	}

	//var s string
	// If the value is a quoted string.
	if data[0] == '"' {
		if !cfg.FromStrings.Allowed {
			return fmt.Errorf("dirty.Number: string input not allowed")
		}
		if len(data) < 2 || data[len(data)-1] != '"' {
			return errors.New("dirty.Number: invalid string value")
		}
		s := string(data[1 : len(data)-1])
		s = strings.TrimSpace(s)

		// Remove spaces if allowed.
		if cfg.FromStrings.SpacingAllowed {
			s = strings.ReplaceAll(s, " ", "")
		}
		// Remove commas if allowed.
		if cfg.FromStrings.CommasAllowed {
			s = strings.ReplaceAll(s, ",", "")
		}

		// TODO: ensure cfg.FromStrings.ExponentNotationAllowed is respected

		// Parse the float.
		n, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			return fmt.Errorf("dirty.Number: cannot parse number: %w", err)
		}

		// TODO: handle cfg.FromStrings.FloatishAllowed
		// we can't know about it here, as we don't know the destination clean type
		// (and we probably won't never know it here. so it will be at a later stage)

		*v = Number(n)
		return nil
	}

	// Raw token (can be number, boolean, null, objet, array)
	s := strings.TrimSpace(string(data))

	if s[0] == 'n' /* null  */ {
		if cfg.FromNull.Allowed {
			*v = Number(0.0)
			return nil
		}

		return errors.New("dirty.Number: numbers from nulls are not allowed")
	} else if s[0] == 't' {
		if cfg.FromBools.Allowed {
			*v = Number(1.0)
			return nil
		}
		return errors.New("dirty.Number: numbers from bools are not allowed")
	} else if s[0] == 'f' {
		if cfg.FromBools.Allowed {
			*v = Number(0.0)
			return nil
		}
		return errors.New("dirty.Number: numbers from bools are not allowed")
	} else if s[0] == '[' || s[0] == '{' {
		return errors.New("dirty.Number: can't parse bools from object/array values")
	}

	// should be a regular number-ish value.

	// Parse the float.
	n, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
	if err != nil {
		return fmt.Errorf("dirty.Number: cannot parse number: %w", err)
	}
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
		return fmt.Errorf("dirty.String can't be parsed from: %s", s)
	}
	if len(data) < 2 || data[len(data)-1] != '"' {
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

	cfg := config.Global().Bool
	if !cfg.Allowed { // TODO: that's bad. do it better
		cfg.FromNull.Allowed = false
		cfg.FromStrings.Allowed = false
		cfg.FromNull.Allowed = false
	}

	var (
		boolFromNumber = func(n float64) option.Bool {

			var b option.Bool
			if parser, ok := parsersBoolFromNum[cfg.FromNumbers.CustomParseFunc]; ok {
				b = parser(n)
			} else {
				// TRICKY THING. CORRUPTED CONFIG IS HERE. We should not just silenty exit
				// Let's log or something similar (TODO: handle this carefully)
				return option.NoneBool()
			}

			if b.Some() {
				return b
			}

			if cfg.FromNumbers.FallbackValue.Some() {
				return cfg.FromNumbers.FallbackValue
			}

			return option.NoneBool()
		}

		boolFromString = func(s string) option.Bool {
			// Check against the configured true strings.
			for _, ts := range cfg.FromStrings.CustomListForTrue {
				if s == strings.ToLower(ts) {
					return option.True()
				}
			}
			// Check against the configured false strings.
			for _, fs := range cfg.FromStrings.CustomListForFalse {
				if s == strings.ToLower(fs) {
					return option.False()
				}
			}

			if cfg.FromStrings.RespectFromNumbersLogic {
				if v, err := strconv.ParseFloat(s, 64); err == nil {
					return boolFromNumber(v)
				}
			}

			return cfg.FromStrings.FallbackValue
		}
	)

	// Check if the incoming value is a quoted string.
	if data[0] == '"' {
		if !cfg.FromStrings.Allowed {
			return fmt.Errorf("dirty.Bool: string input not allowed")
		}

		// Valid strings are considered to be quoted from both sides
		if len(data) < 2 || data[len(data)-1] != '"' {
			return errors.New("dirty.Bool: corrupt string value")
		}
		s := string(data[1 : len(data)-1])
		s = strings.TrimSpace(strings.ToLower(s)) // normalized content of the string

		if s == "" && cfg.FromStrings.FalseForEmptyString {
			*v = false
			return nil
		}

		if b := boolFromString(s); b.Some() {
			*v = Bool(b.Unwrap())
			return nil
		}

		return fmt.Errorf("dirty.Bool: cannot parse string (%q) as bool", limitedStr(s, 50))
	}

	// Raw token (can be number, boolean, or anything else)

	s := string(data)

	// As we consider it a valid JSON, if first letter is `t` or `f` then it definetely true/false
	if s[0] == 't' {
		*v = true
		return nil
	} else if s[0] == 'f' {
		*v = false
		return nil
	} else if s[0] == 'n' /* null  */ {
		if cfg.FromNull.Allowed {
			*v = Bool(cfg.FromNull.Inverse) // if Inverse: we'll return true, otherwise: false
		}

		return errors.New("dirty.Bool: cannot parse bool from null")
	}

	if s[0] == '{' || s[0] == '[' {
		return errors.New("dirty.Bool: can't parse bools from object/array values")
	}

	// Should be a number then
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("dirty.Bool: cannot parse as bool (%q): %w", limitedStr(s, 50), err)
	}

	if b := boolFromNumber(n); b.Some() {
		*v = Bool(b.Unwrap())
		return nil
	}

	return fmt.Errorf("dirty.Bool: unrecognized value for bool (%q)", limitedStr(s, 50))
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

func limitedStr(s string, limit int) string {
	if len(s) > limit {
		return s[0:limit] + "…"
	}

	return s
}
