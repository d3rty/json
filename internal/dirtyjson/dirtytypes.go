package dirtyjson

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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

// Dirtyable is used as a way to link a clean model with a dirty model.
type Dirtyable interface {
	Dirty() any
}

// Enabled is a struct atom that enables dirty unmarshalling for the struct where it's embedded.
type Enabled struct {
	res any
}

func (e *Enabled) result() any { return e.res }
func (e *Enabled) init(v any)  { e.res = v }

// Disabled is an atom struct that that remains a syntactically valid dirty model
// but disables dirty unmarshalling.
// You can easily switch from `dirty.Enabled` to `dirty.Disabled`
// keeping all models and interfaces working (falling back to standard (clean) json.Unmarshal).
type Disabled struct{}

func (*Disabled) result() any { return nil }
func (*Disabled) init(_ any)  {}
func (*Disabled) isDisabled() {} // isDisabled disabled dirtying (keeping all interfaces working)

type (
	// Number means any number (Integer, Float, Scientific, etc.)
	Number float64
	// String means simply a string.
	String string
	// Bool means a boolean value.
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

	// Time means a given moment within a day (day is not specified in the source value).
	Time time.Time

	// TODO(github.com/d3rty/json/issues/4): SmartScalar
	// TODO(github.com/d3rty/json/issues/5): StringifiedContainers
)

const (
	literalForTrue  = "true"
	literalForFalse = "false"
)

// getConfig returns the config from the given context.
func getConfig(ctx context.Context) *config.Config {
	_ = ctx
	// TODO(github.com/d3rty/json/issues/6): support for per-ctx configs
	return config.Global()
}

// UnmarshalJSON converts []byte into a Number.
func (v *Number) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("dirty.Number: UnmarshalJSON on nil pointer")
	}

	fullCfg := getConfig(context.Background())

	if fullCfg.Number.IsDisabled() { // Dirty number decoding is disabled
		var clean float64
		if err := json.Unmarshal(data, &clean); err != nil {
			return err
		}
		*v = Number(clean)
		return nil
	}

	// cfg stays for specifical config of Number decoding
	cfg := fullCfg.Number

	// var s string
	// If the value is a quoted string.
	if data[0] == '"' {
		if cfg.FromStrings.IsDisabled() {
			return errors.New("dirty.Number: string input not allowed")
		}
		s, err := getStringBetweenQuotes(data)
		if err != nil {
			return errors.New("dirty.Number: invalid string value")
		}

		fromStringsCfg := cfg.FromStrings

		// Remove spaces if allowed.
		if fromStringsCfg.SpacingAllowed {
			s = strings.ReplaceAll(s, " ", "")
		}
		// Remove commas if allowed.
		if fromStringsCfg.CommasAllowed {
			s = strings.ReplaceAll(s, ",", "")
		}

		// TODO(github.com/d3rty/json/issues/10): respect all things from cfg.FromStrings
		// 		Note on handle cfg.FromStrings.FloatishAllowed:
		//      we can't know about it here, as we don't know the destination clean type
		//		(and we probably won't never know it here. so it will be at a later stage)

		// Parse the float.
		n, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			return fmt.Errorf("dirty.Number: cannot parse number: %w", err)
		}

		*v = Number(n)
		return nil
	}

	// Raw token (can be number, boolean, null, object, array)
	s := strings.TrimSpace(string(data))

	switch s[0] {
	case 'n': /* null  */
		if cfg.FromNull.IsDisabled() {
			return errors.New("dirty.Number: numbers from nulls are not allowed")
		}
		*v = Number(0.0)
		return nil

	case 't':
		if cfg.FromBools.IsDisabled() {
			return errors.New("dirty.Number: numbers from bools are not allowed")
		}
		*v = Number(1.0)
		return nil

	case 'f':
		if cfg.FromBools.IsDisabled() {
			return errors.New("dirty.Number: numbers from bools are not allowed")
		}
		*v = Number(0.0)
		return nil

	case '[', '{':
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

// UnmarshalJSON converts []byte into an Integer.
func (v *Integer) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("dirty.Integer: UnmarshalJSON on nil pointer")
	}

	fullCfg := getConfig(context.Background())

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

		s, err := getStringBetweenQuotes(data)
		if err != nil {
			return errors.New("dirty.Integer: invalid string value")
		}

		// Remove spaces if allowed.
		if fromStringsCfg.SpacingAllowed {
			s = strings.ReplaceAll(s, " ", "")
		}
		// Remove commas if allowed.
		if fromStringsCfg.CommasAllowed {
			s = strings.ReplaceAll(s, ",", "")
		}

		// TODO(github.com/d3rty/json/issues/10): respect all things from cfg.FromStrings
		// 		Note on handle cfg.FromStrings.FloatishAllowed:
		//      we can't know about it here, as we don't know the destination clean type
		//		(and we probably won't never know it here. so it will be at a later stage)

		// Parse the float.
		n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if err != nil {
			return fmt.Errorf("dirty.Number: cannot parse number: %w", err)
		}

		*v = Integer(n)
		return nil
	}

	// Raw token (can be number, boolean, null, object, array)
	s := strings.TrimSpace(string(data))

	switch s[0] {
	case 'n': /* null  */
		if cfg.FromNull.IsDisabled() {
			return errors.New("dirty.Integer: numbers from nulls are not allowed")
		}
		*v = Integer(0)
		return nil

	case 't':
		if cfg.FromBools.IsDisabled() {
			return errors.New("dirty.Integer: numbers from bools are not allowed")
		}
		*v = Integer(1)
		return nil

	case 'f':
		if cfg.FromBools.IsDisabled() {
			return errors.New("dirty.Integer: numbers from bools are not allowed")
		}
		*v = Integer(0)
		return nil

	case '[', '{':
		return errors.New("dirty.Integer: can't parse bools from object/array values")
	}

	// should be a regular integer value.

	// Parse the float.
	// TODO(github.com/d3rty/json/issues/12) Roundable floats?
	n, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return fmt.Errorf("dirty.Integer: cannot parse number: %w", err)
	}
	*v = Integer(n)
	return nil
}

// UnmarshalJSON converts []byte into a Bool.
func (v *String) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("dirty.String: UnmarshalJSON on nil pointer")
	}

	s, err := getStringBetweenQuotes(data)
	if err != nil {
		return errors.New("dirty.String: invalid string value")
	}

	*v = String(s)
	return nil
}

// UnmarshalJSON converts []byte into a Bool.
func (v *Bool) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("dirty.Bool: UnmarshalJSON on nil pointer")
	}

	fullCfg := getConfig(context.Background())

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
				// TRICKY THING. CORRUPTED CONFIG IS HERE.
				// TODO(github.com/d3rty/json/issues/11): do a loud log here
				slog.Error("possible corrupted config ", "parse_func", fromNumbersCfg.CustomParseFunc)
				return option.NoneBool()
			}

			if b.Some() {
				return b
			}

			return cfg.FallbackValue
		}

		boolFromString = func(s string, cfg *config.BoolFromStringsConfig) option.Bool {
			if s == "" {
				// if not presented in custom lists, then assume it as false
				if !slices.Contains(cfg.CustomListForTrue, "") && !slices.Contains(cfg.CustomListForFalse, "") {
					return option.False()
				}

				// otherwise continue with regular logic
			}

			sLower := strings.ToLower(s)

			// handling true via CustomListForTrue (or literalForTrue):
			if len(cfg.CustomListForTrue) > 0 {
				if cfg.CaseInsensitive {
					for _, ts := range cfg.CustomListForTrue {
						if sLower == strings.ToLower(ts) {
							return option.True()
						}
					}
				} else if slices.Contains(cfg.CustomListForTrue, s) {
					return option.True()
				}
			} else {
				if literalForTrue == s {
					return option.True()
				}
				if cfg.CaseInsensitive && literalForTrue == strings.ToLower(s) {
					return option.True()
				}
			}

			if len(cfg.CustomListForFalse) > 0 {
				if cfg.CaseInsensitive {
					for _, ts := range cfg.CustomListForFalse {
						if sLower == strings.ToLower(ts) {
							return option.False()
						}
					}
				} else if slices.Contains(cfg.CustomListForFalse, s) {
					return option.False()
				}
			} else {
				if literalForFalse == s {
					return option.False()
				}
				if cfg.CaseInsensitive && literalForFalse == strings.ToLower(s) {
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

		s, err := getStringBetweenQuotes(data)
		if err != nil {
			return errors.New("dirty.Bool: invalid string value")
		}

		cfgFromStrings := cfg.FromStrings

		if b := boolFromString(s, cfgFromStrings); b.Some() {
			*v = Bool(b.Unwrap())
			return nil
		}

		return fmt.Errorf("dirty.Bool: cannot parse string (%q) as bool", limitedStr(s))
	}

	// Raw token (can be number, boolean, or anything else)

	s := string(data)

	// As we consider it a valid JSON, if the first letter is 't' or 'f' then it definitely true/false
	switch s[0] {
	case 't':
		*v = true
		return nil
	case 'f':
		*v = false
		return nil
	case 'n': /* null  */
		if cfg.FromNull.IsDisabled() {
			return errors.New("dirty.Bool: cannot parse bool from null")
		}
		*v = Bool(cfg.FromNull.Inverse) // if Inverse: we'll return true, otherwise: false
		return nil
	case '{', '[':
		return errors.New("dirty.Bool: can't parse bools from object/array values")
	}

	// Should be a number then
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("dirty.Bool: cannot parse as bool (%q): %w", limitedStr(s), err)
	}

	if b := boolFromNumber(n); b.Some() {
		*v = Bool(b.Unwrap())
		return nil
	}

	return fmt.Errorf("dirty.Bool: unrecognized value for bool (%q)", limitedStr(s))
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
	*v = arr
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

// UnmarshalJSON converts []byte into a Date.
func (v *DateTime) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("dirty.DateTime: UnmarshalJSON on nil pointer")
	}

	var err error
	*v, err = unmarshalDateTime[DateTime](context.Background(), data)
	return err
}

// UnmarshalJSON converts []byte into a Date.
func (v *Date) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("dirty.Date: UnmarshalJSON on nil pointer")
	}

	res, err := unmarshalDateTime[Date](context.Background(), data)
	if err != nil {
		return err
	}

	// trimming DateTime to Date
	var t = time.Time(res)
	*v = Date(
		years.Mutate(&t).TruncateToDay().Time(),
	)
	return nil
}

// UnmarshalJSON converts []byte into a Date.
func (v *Time) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("dirty.Date: UnmarshalJSON on nil pointer")
	}

	res, err := unmarshalDateTime[Time](context.Background(), data)
	if err != nil {
		return fmt.Errorf("dirty.Date unmarshal failure: %w", err)
	}

	var t = time.Time(res)
	*v = Time(
		years.Mutate(&t).SetYear(0).SetMonth(0).SetDay(0).Time(),
	)
	return nil
}

func unmarshalDateTime[T Date | DateTime | Time](ctx context.Context, data []byte) (T, error) {
	fullCfg := getConfig(ctx)

	var zero T

	if fullCfg.Date.IsDisabled() { // Dirty date decoding is disabled
		var clean T
		if err := json.Unmarshal(data, &clean); err != nil {
			return zero, err
		}

		return clean, nil
	}

	cfg := fullCfg.Date

	// If the value is a quoted string.
	if data[0] == '"' {
		if cfg.FromStrings.IsDisabled() {
			return zero, errors.New("dirty.DateTime: string input not allowed")
		}

		s, err := getStringBetweenQuotes(data)
		if err != nil {
			return zero, errors.New("dirty.DateTime: invalid string value")
		}

		opts := make([]years.ParserOption, 0)
		if cfg.FromStrings.Aliases {
			opts = append(opts, years.AcceptAliases())
		}

		var layouts []string
		switch any(zero).(type) {
		case DateTime:
			layouts = cfg.FromStrings.Layouts.DateTime
		case Date:
			layouts = cfg.FromStrings.Layouts.Date
		case Time:
			layouts = cfg.FromStrings.Layouts.Date
		default:
			panic("unhandled date format")
		}

		if len(layouts) > 0 {
			opts = append(opts, years.WithLayouts(layouts...))
		}
		if cfg.FromStrings.RespectFromNumbersLogic && !cfg.FromNumbers.IsDisabled() {
			if cfg.FromNumbers.UnixTimestamp {
				opts = append(opts, years.AcceptUnixSeconds())
			}
			if cfg.FromNumbers.UnixMilliTimestamp {
				opts = append(opts, years.AcceptUnixMilli())
			}
		}

		parsed, err := years.NewParser(opts...).JustParse(s)
		if err != nil {
			return zero, errors.New("dirty.DateTime: couldn't parse datetime value")
		}

		return T(parsed), nil
	}

	// Raw token (can be number, null, objet, array)
	s := strings.TrimSpace(string(data))

	switch s[0] {
	case 'n': /* null  */
		if cfg.FromNull.IsDisabled() {
			return zero, errors.New("dirty.DateTime: dates from nulls are not allowed")
		}
		return zero, nil
	case 't', 'f':
		return zero, errors.New("dirty.DateTime: can't parse dates from boolean values")
	case '[', '{':
		return zero, errors.New("dirty.DateTime: can't parse dates from object/array values")
	}

	// should be a regular integer value.
	if cfg.FromNumbers.IsDisabled() {
		return zero, errors.New("dirty.DateTime: dates from numbers are not allowed")
	}

	opts := make([]years.ParserOption, 0)
	if cfg.FromNumbers.UnixTimestamp {
		opts = append(opts, years.AcceptUnixSeconds())
	}
	if cfg.FromNumbers.UnixMilliTimestamp {
		opts = append(opts, years.AcceptUnixMilli())
	}

	parsed, err := years.NewParser(opts...).JustParse(s)
	if err != nil {
		return zero, fmt.Errorf("dirty.DateTime: cannot parse numeric date: %w", err)
	}
	return T(parsed), nil
}
