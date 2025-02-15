package dirty

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type (
	// Number is a custom type for unmarshalling numbers.
	// Numbers can be parsed from strings or actual JSON numbers.
	// Other types will be rejected.
	Number float64

	// String s a custom type for unmarshalling strings.
	// Anything except actual json string will be rejected.
	String string

	// Bool is a custom type for unmarshalling booleans.
	// Bools can be parsed from
	// 	- strings ("true", "false", "yes", "no", "on", "off", "1", "0")
	//  - numbers (1, 0)
	//  - actual JSON booleans.
	// Other types will be rejected.
	Bool bool

	// Array is a custom type for unmarshalling arrays.
	// Anything except actual JSON arrays will be rejected.
	Array []any

	// Object is a custom type for unmarshalling objects.
	// Anything except actual JSON objects will be rejected.
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
	*v = Object(obj)
	return nil
}
