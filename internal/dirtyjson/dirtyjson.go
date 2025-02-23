package dirtyjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
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

	var s string
	switch data[0] {
	case '"':
		if len(data) < 2 {
			return errors.New("dirty.Number missing closing quote")
		}
		if data[len(data)-1] != '"' {
			return errors.New("dirty.Number missing closing quote")
		}

		s = string(data[1 : len(data)-1])
	default:
		s = string(data)
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("dirty.Number can't parse a number: %w", err)
	}
	*v = Number(f)
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

	var s string
	if data[0] == '"' {
		if len(data) < 2 {
			return errors.New("dirty.Bool missing closing quote")
		}
		if data[len(data)-1] != '"' {
			return errors.New("dirty.Bool missing closing quote")
		}

		s = string(data[1 : len(data)-1])

		switch strings.ToLower(s) {
		case "true", "yes", "on", "1":
			*v = true
			return nil
		case "false", "no", "off", "0":
			*v = false
			return nil
		default:
			return fmt.Errorf("dirty.Bool cant be parsed from json content: %s", s)
		}
	}

	s = string(data)

	if s == "true" {
		*v = true
		return nil
	} else if s == "false" {
		*v = false
		return nil
	}

	return fmt.Errorf("dirty.Bool cant be parsed from: %s", s)
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
