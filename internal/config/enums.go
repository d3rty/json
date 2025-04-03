package config

import (
	"fmt"
	"strings"
)

// undefined is a constant string used for undefined values of any enum.
const undefined = ""

// BoolFromNumberAlg specifies the algorithm of how parsing Number->Bool is done.
type BoolFromNumberAlg uint8

const (
	// BoolFromNumberUndefined is the undefined value.
	BoolFromNumberUndefined BoolFromNumberAlg = 0

	// BoolFromNumberBinary is the "1/0" parser. 1 is true, 0 is false.
	// Other numbers are considered "non parsed" (fallback value or Red result).
	BoolFromNumberBinary BoolFromNumberAlg = 1 << (iota - 1) // 1 (001)

	// BoolFromNumberPositiveNegative is the "<=0 vs >0" parser.
	// Positive numbers are true. Negative numbers And zero are false.
	BoolFromNumberPositiveNegative // 2 (010)

	// BoolFromNumberSignOfOne is the "-1/1" parser.
	// -1 means false, 1 means true. Other numbers are considerd "non parsed" (fallback value or Red result).
	BoolFromNumberSignOfOne // 4 (100)
)

// enumBoolFromNumberAlgs stores the enum string -> Value.
//
//nolint:gochecknoglobals // we're OK with it
var enumBoolFromNumberAlgs = map[string]BoolFromNumberAlg{
	"binary":            BoolFromNumberBinary,
	"positive_negative": BoolFromNumberPositiveNegative,
	"sign_of_one":       BoolFromNumberSignOfOne,
	undefined:           BoolFromNumberUndefined,
}

func (*BoolFromNumberAlg) DefaultValue() BoolFromNumberAlg { return BoolFromNumberBinary }

// UnmarshalText implements the encoding.TextUnmarshaler interface for BoolFromNumberAlg.
// It converts a string (e.g., "sign_of_one") into the corresponding enum value.
func (b *BoolFromNumberAlg) UnmarshalText(text []byte) error {
	s := strings.ToLower(strings.TrimSpace(string(text)))
	if v, ok := enumBoolFromNumberAlgs[s]; ok {
		*b = v
		return nil
	}

	return fmt.Errorf("unknown BoolFromNumberAlg value: %q", s)
}

// MarshalText implements the encoding.TextMarshaler interface for BoolFromNumberAlg.
// It converts enum value into its string representation.
func (b *BoolFromNumberAlg) MarshalText() ([]byte, error) {
	if s := b.String(); s != undefined {
		return []byte(s), nil
	}

	return []byte(undefined), nil
}

// String stringifies value of BoolFromNumberAlg.
func (b *BoolFromNumberAlg) String() string {
	if b != nil {
		for s, v := range enumBoolFromNumberAlgs {
			if v == *b {
				return s
			}
		}
	}

	return undefined
}

// ListAvailableBoolFromNumberAlgs lists all available values of BoolFromNumberAlg.
func ListAvailableBoolFromNumberAlgs() []BoolFromNumberAlg {
	algs := make([]BoolFromNumberAlg, 0, len(enumBoolFromNumberAlgs)-1)

	for _, alg := range enumBoolFromNumberAlgs {
		if alg == BoolFromNumberUndefined {
			continue
		}

		algs = append(algs, alg)
	}

	return algs
}

// RoundingAlg specifies the algorithm of how parsing Number->Bool is done.
type RoundingAlg uint8

const (
	// RoundingAlgUndefined is the undefined value.
	RoundingAlgUndefined RoundingAlg = 0

	// RoundingAlgNone means integers can't be parsed from floors with non-zero decimals.
	RoundingAlgNone RoundingAlg = 1 << (iota - 1) // 1 (001)

	// RoundingAlgFloor means it uses math.Floor() when parsing integers from floats.
	RoundingAlgFloor // 2 (010)

	// RoundingAlgRound means it uses math.Round() when parsing integers from floats.
	RoundingAlgRound // 4 (100)
)

// enumRoundingAlgs stores the enum string -> Value.
//
//nolint:gochecknoglobals // we're OK with it
var enumRoundingAlgs = map[string]RoundingAlg{
	"none":    RoundingAlgNone,
	"floor":   RoundingAlgFloor,
	"round":   RoundingAlgRound,
	undefined: RoundingAlgUndefined,
}

func (*RoundingAlg) DefaultValue() RoundingAlg { return RoundingAlgNone }

// UnmarshalText implements the encoding.TextUnmarshaler interface for RoundingAlg.
// It converts a string (e.g., "floor") into the corresponding enum value.
func (b *RoundingAlg) UnmarshalText(text []byte) error {
	s := strings.ToLower(strings.TrimSpace(string(text)))
	if v, ok := enumRoundingAlgs[s]; ok {
		*b = v
		return nil
	}

	return fmt.Errorf("unknown enumRoundingAlgs value: %q", s)
}

// MarshalText implements the encoding.TextMarshaler interface for RoundingAlg.
// It converts enum value into its string representation.
func (b *RoundingAlg) MarshalText() ([]byte, error) {
	if s := b.String(); s != undefined {
		return []byte(s), nil
	}

	return []byte(undefined), nil
}

// String stringifies value of RoundingAlg.
func (b *RoundingAlg) String() string {
	if b != nil {
		for s, v := range enumRoundingAlgs {
			if v == *b {
				return s
			}
		}
	}

	return undefined
}

// ListAvailableRoundingAlgs lists all available values of RoundingAlg.
func ListAvailableRoundingAlgs() []RoundingAlg {
	algs := make([]RoundingAlg, 0, len(enumRoundingAlgs)-1)

	for _, alg := range enumRoundingAlgs {
		if alg == RoundingAlgUndefined {
			continue
		}

		algs = append(algs, alg)
	}

	return algs
}
