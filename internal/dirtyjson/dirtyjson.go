package dirtyjson

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/d3rty/json/internal/config"
	"github.com/d3rty/json/internal/option"

	"github.com/amberpixels/years"
)

// that allow us to init and retrieve dirty data.
type d3rtyContainer interface {
	init(any)
	result() any
}

// It's used as a way to link clean model with dirty model.
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
	// Number means any number (Integer, Float, Scientific, etc.)
	Number float64
	// String means simply a string.
	String string
	// Bool meansns a boolean value.
	Bool bool
	// Array means array of anything.
	Array []any
	// Object is a JSON-like map (string->any).
	Object map[string]any

	// Integer means an Integer number.
	Integer int64

	// Date means a date (time pointing to a specific day).
	Date time.Time

	// DateTime means a time (time pointing to a specific moment).
	DateTime time.Time

	// SmartScalar means that it respects the type of given value
	// For Bools and Floats: it remains bool or float64.
	// For Null: it remains nil.
	// For String:
	// 		if it's numerish string - it will be float64
	//     	if it's boolish ("true"/"false") string - it will be a Bool
	//     	otherwise it's a string
	//
	// For non-scalar value - it can't be parsed.
	//
	// TODO: allow config to setup how flexible the numerish/boolish strings are.
	SmartScalar struct {
		scalar any
	}

	// TODO: Arrays from String, Objects from strings. When some part of nested JSON is stringifed.
)

// TODO(?) non-global config.
func tmpGetConfig(ctx context.Context) *config.Config {
	_ = ctx
	return config.Global()
}

// UnmarshalJSON converts []byte into a Number.
func (v *Number) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("dirty.Number: UnmarshalJSON on nil pointer")
	}

	fullCfg := tmpGetConfig(context.Background())

	if fullCfg.Number.IsDisabled() { // Dirty number decoding is disabled
		var clean float64
		if err := json.Unmarshal(data, &clean); err != nil {
			return err
		}
		*v = Number(clean)
		return nil
	}

	// cfg stays for specifically config of Number decoding
	cfg := fullCfg.Number

	// var s string
	// If the value is a quoted string.
	if data[0] == '"' {
		if cfg.FromStrings.IsDisabled() {
			return errors.New("dirty.Number: string input not allowed")
		}
		if len(data) < 2 || data[len(data)-1] != '"' {
			return errors.New("dirty.Number: invalid string value")
		}
		s := string(data[1 : len(data)-1])
		s = strings.TrimSpace(s)

		fromStringsCfg := cfg.FromStrings

		// Remove spaces if allowed.
		if fromStringsCfg.SpacingAllowed {
			s = strings.ReplaceAll(s, " ", "")
		}
		// Remove commas if allowed.
		if fromStringsCfg.CommasAllowed {
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

	switch {
	case s[0] == 'n': /* null  */
		if cfg.FromNull.IsDisabled() {
			return errors.New("dirty.Number: numbers from nulls are not allowed")
		}
		*v = Number(0.0)
		return nil

	case s[0] == 't':
		if cfg.FromBools.IsDisabled() {
			return errors.New("dirty.Number: numbers from bools are not allowed")
		}
		*v = Number(1.0)
		return nil

	case s[0] == 'f':
		if cfg.FromBools.IsDisabled() {
			return errors.New("dirty.Number: numbers from bools are not allowed")
		}
		*v = Number(0.0)
		return nil

	case s[0] == '[' || s[0] == '{':
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
//
//nolint:funlen // we're OK
func (v *Bool) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("dirty.Bool: UnmarshalJSON on nil pointer")
	}

	fullCfg := tmpGetConfig(context.Background())

	if fullCfg.Bool.IsDisabled() { // Dirty bool decoding is disabled
		var clean bool
		if err := json.Unmarshal(data, &clean); err != nil {
			return err
		}
		*v = Bool(clean)
		return nil
	}

	cfg := fullCfg.Bool

	var (
		boolFromNumber = func(n float64) option.Bool {
			// assuming config is enabled
			fromNumbersCfg := cfg.FromNumbers
			var b option.Bool
			if parser, ok := parsersBoolFromNum[fromNumbersCfg.CustomParseFunc]; ok {
				b = parser(n)
			} else {
				// TRICKY THING. CORRUPTED CONFIG IS HERE. We should not just silenty exit
				// Let's log or something similar (TODO: handle this carefully)
				return option.NoneBool()
			}

			if b.Some() {
				return b
			}

			return cfg.FallbackValue
		}

		boolFromString = func(s string, cfg *config.BoolFromStringsConfig) option.Bool {
			sLower := strings.ToLower(s)

			if cfg.CaseInsensitive {
				for _, ts := range cfg.CustomListForTrue {
					if sLower == strings.ToLower(ts) {
						return option.True()
					}
				}
				for _, fs := range cfg.CustomListForFalse {
					if sLower == strings.ToLower(fs) {
						return option.False()
					}
				}
			} else {
				if slices.Contains(cfg.CustomListForTrue, s) {
					return option.True()
				}
				if slices.Contains(cfg.CustomListForFalse, s) {
					return option.False()
				}
			}

			if cfg.RespectFromNumbersLogic {
				if v, err := strconv.ParseFloat(s, 64); err == nil {
					return boolFromNumber(v)
				}
			}

			return fullCfg.Bool.FallbackValue
		}
	)

	// Check if the incoming value is a quoted string.
	if data[0] == '"' {
		if cfg.FromStrings.IsDisabled() {
			return errors.New("dirty.Bool: string input not allowed")
		}

		// Valid strings are considered to be quoted from both sides
		if len(data) < 2 || data[len(data)-1] != '"' {
			return errors.New("dirty.Bool: corrupt string value")
		}
		s := string(data[1 : len(data)-1])
		s = strings.TrimSpace(s) // normalized content of the string

		cfgFromStrings := cfg.FromStrings

		if s == "" && cfgFromStrings.FalseForEmptyString {
			*v = false
			return nil
		}

		if b := boolFromString(s, cfgFromStrings); b.Some() {
			*v = Bool(b.Unwrap())
			return nil
		}

		return fmt.Errorf("dirty.Bool: cannot parse string (%q) as bool", limitedStr(s, maxMessageLength))
	}

	// Raw token (can be number, boolean, or anything else)

	s := string(data)

	// As we consider it a valid JSON, if first letter is `t` or `f` then it definetely true/false
	switch {
	case s[0] == 't':
		*v = true
		return nil
	case s[0] == 'f':
		*v = false
		return nil
	case s[0] == 'n': /* null  */
		if cfg.FromNull.IsDisabled() {
			return errors.New("dirty.Bool: cannot parse bool from null")
		}
		*v = Bool(cfg.FromNull.Inverse) // if Inverse: we'll return true, otherwise: false
		return nil
	}

	if s[0] == '{' || s[0] == '[' {
		return errors.New("dirty.Bool: can't parse bools from object/array values")
	}

	// Should be a number then
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("dirty.Bool: cannot parse as bool (%q): %w", limitedStr(s, maxMessageLength), err)
	}

	if b := boolFromNumber(n); b.Some() {
		*v = Bool(b.Unwrap())
		return nil
	}

	return fmt.Errorf("dirty.Bool: unrecognized value for bool (%q)", limitedStr(s, maxMessageLength))
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

// UnmarshalJSON converts []byte into an Integer.
func (v *Integer) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("dirty.Integer: UnmarshalJSON on nil pointer")
	}

	fullCfg := tmpGetConfig(context.Background())

	if fullCfg.Number.IsDisabled() { // Dirty number decoding is disabled
		var clean int64
		if err := json.Unmarshal(data, &clean); err != nil {
			return err
		}

		*v = Integer(clean)
		return nil
	}

	cfg := fullCfg.Number

	// If the value is a quoted string.
	if data[0] == '"' {
		if cfg.FromStrings.IsDisabled() {
			return errors.New("dirty.Integer: string input not allowed")
		}
		fromStringsCfg := cfg.FromStrings

		if len(data) < 2 || data[len(data)-1] != '"' {
			return errors.New("dirty.Integer: invalid string value")
		}
		s := string(data[1 : len(data)-1])
		s = strings.TrimSpace(s)

		// Remove spaces if allowed.
		if fromStringsCfg.SpacingAllowed {
			s = strings.ReplaceAll(s, " ", "")
		}
		// Remove commas if allowed.
		if fromStringsCfg.CommasAllowed {
			s = strings.ReplaceAll(s, ",", "")
		}

		// TODO: ensure cfg.FromStrings.ExponentNotationAllowed is respected

		// Parse the float.
		n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if err != nil {
			return fmt.Errorf("dirty.Number: cannot parse number: %w", err)
		}

		// TODO: handle cfg.FromStrings.FloatishAllowed
		// we can't know about it here, as we don't know the destination clean type
		// (and we probably won't never know it here. so it will be at a later stage)

		*v = Integer(n)
		return nil
	}

	// Raw token (can be number, boolean, null, objet, array)
	s := strings.TrimSpace(string(data))

	switch {
	case s[0] == 'n': /* null  */
		if cfg.FromNull.IsDisabled() {
			return errors.New("dirty.Integer: numbers from nulls are not allowed")
		}
		*v = Integer(0)
		return nil

	case s[0] == 't':
		if cfg.FromBools.IsDisabled() {
			return errors.New("dirty.Integer: numbers from bools are not allowed")
		}
		*v = Integer(1)
		return nil

	case s[0] == 'f':
		if cfg.FromBools.IsDisabled() {
			return errors.New("dirty.Integer: numbers from bools are not allowed")
		}
		*v = Integer(0)
		return nil

	case s[0] == '[' || s[0] == '{':
		return errors.New("dirty.Integer: can't parse bools from object/array values")
	}

	// should be a regular integer value.

	// Parse the float.
	// TODO: configurable: if we allow to "round" floats??
	n, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return fmt.Errorf("dirty.Integer: cannot parse number: %w", err)
	}
	*v = Integer(n)
	return nil
}

// UnmarshalJSON converts []byte into an Date.
//
//nolint:dupl // it's not a real dupl TODO can we reuse the code here?
func (v *Date) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("dirty.Date: UnmarshalJSON on nil pointer")
	}

	fullCfg := tmpGetConfig(context.Background())

	if fullCfg.Date.IsDisabled() { // Dirty date decoding is disabled
		var clean time.Time
		if err := json.Unmarshal(data, &clean); err != nil {
			return err
		}

		*v = Date(clean)
		return nil
	}

	cfg := fullCfg.Date
	// If the value is a quoted string.
	if data[0] == '"' {
		if cfg.FromStrings.IsDisabled() {
			return errors.New("dirty.Date: string input not allowed")
		}
		if len(data) < 2 || data[len(data)-1] != '"' {
			return errors.New("dirty.Date: invalid string value")
		}
		s := string(data[1 : len(data)-1])
		s = strings.TrimSpace(s)

		parsed, err := years.JustParse(s)
		if err != nil {
			return errors.New("dirty.Date: couldn't parse datetime value")
		}

		*v = Date(parsed)
		return nil
	}

	// Raw token (can be number, null, objet, array)
	s := strings.TrimSpace(string(data))

	switch {
	case s[0] == 'n': /* null  */
		if cfg.FromNull.IsDisabled() {
			return errors.New("dirty.Date: dates from nulls are not allowed")
		}

		*v = Date(time.Time{}) // only zero time for now
		return nil
	case s[0] == 't' || s[0] == 'f':
		return errors.New("dirty.Date: can't parse dates from boolean values")
	case s[0] == '[' || s[0] == '{':
		return errors.New("dirty.Date: can't parse dates from object/array values")
	}

	// should be a regular integer value.
	if cfg.FromNumbers.IsDisabled() {
		return errors.New("dirty.Date: dates from numbers are not allowed")
	}

	// TODO: respect config
	parsed, err := years.JustParse(strings.TrimSpace(string(data)))
	if err != nil {
		return fmt.Errorf("dirty.Date: cannot parse numeric date: %w", err)
	}
	*v = Date(parsed)
	return nil
}

// UnmarshalJSON converts []byte into an Date.
//
//nolint:dupl // it's not a real dupl TODO can we reuse the code here?
func (v *DateTime) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("dirty.DateTime: UnmarshalJSON on nil pointer")
	}

	fullCfg := tmpGetConfig(context.Background())

	if fullCfg.Date.IsDisabled() { // Dirty date decoding is disabled
		var clean time.Time
		if err := json.Unmarshal(data, &clean); err != nil {
			return err
		}

		*v = DateTime(clean)
		return nil
	}
	cfg := fullCfg.Date

	// If the value is a quoted string.
	if data[0] == '"' {
		if cfg.FromStrings.IsDisabled() {
			return errors.New("dirty.DateTime: string input not allowed")
		}
		if len(data) < 2 || data[len(data)-1] != '"' {
			return errors.New("dirty.DateTime: invalid string value")
		}
		s := string(data[1 : len(data)-1])
		s = strings.TrimSpace(s)

		parsed, err := years.JustParse(s)
		if err != nil {
			return errors.New("dirty.DateTime: couldn't parse datetime value")
		}

		*v = DateTime(parsed)
		return nil
	}

	// Raw token (can be number, null, objet, array)
	s := strings.TrimSpace(string(data))

	switch {
	case s[0] == 'n': /* null  */
		if cfg.FromNull.IsDisabled() {
			return errors.New("dirty.DateTime: dates from nulls are not allowed")
		}
		*v = DateTime(time.Time{}) // only zero time for now
		return nil
	case s[0] == 't' || s[0] == 'f':
		return errors.New("dirty.DateTime: can't parse dates from boolean values")
	case s[0] == '[' || s[0] == '{':
		return errors.New("dirty.DateTime: can't parse dates from object/array values")
	}

	// should be a regular integer value.
	if cfg.FromNumbers.IsDisabled() {
		return errors.New("dirty.DateTime: dates from numbers are not allowed")
	}

	// TODO: respect config
	parsed, err := years.JustParse(strings.TrimSpace(string(data)))
	if err != nil {
		return fmt.Errorf("dirty.DateTime: cannot parse numeric date: %w", err)
	}
	*v = DateTime(parsed)
	return nil
}

// UnmarshalJSON converts []byte into a smart scalar.
func (v *SmartScalar) UnmarshalJSON(data []byte) error {
	if len(data) == 4 {
		if data[0] == 'n' /* null */ {
			v.scalar = nil
			return nil
		}
		if data[0] == 't' /* true */ {
			v.scalar = true
			return nil
		}
	}
	if len(data) == 5 {
		if data[0] == 'f' /* false */ {
			v.scalar = false
			return nil
		}
	}

	// Try unmarshalling as a float64
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		v.scalar = f
		return nil
	}

	// At this point, we assume the data is a JSON string.
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("SmartScalar: unable to unmarshal data: %w", err)
	}

	// If the string is "true" or "false", interpret as bool.
	if s == "true" || s == "false" {
		v.scalar = s == "true"
		return nil
	}

	// If the string can be parsed as a number, interpret as float64.
	if num, err := strconv.ParseFloat(s, 64); err == nil {
		v.scalar = num
		return nil
	}

	// Otherwise, leave it as a string.
	v.scalar = s
	return nil
}

// MarshalJSON unwraps the underlying value and marshals it.
func (v SmartScalar) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.scalar)
}

const (
	maxMessageLength = 50
)

func limitedStr(s string, limit int) string {
	if len(s) > limit {
		return s[0:limit] + "…"
	}

	return s
}
