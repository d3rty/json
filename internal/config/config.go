package config

import (
	"encoding/json"
	"sync"

	"github.com/d3rty/json/internal/option"
)

type BoolFromNumberParser string

const (
	// BoolFromNumberParserBinary is the "1/0" parser. 1 is true, 0 is false.
	// Other numbers are considerd "non parsed" (fallback value or Red result).
	BoolFromNumberBinary BoolFromNumberParser = "binary"

	// BoolFromNumberParserPositiveNegative is the "<=0 vs >0" parser.
	// Positive numbers are true. Negative numbers And zero are false.
	BoolFromNumberPositiveNegative BoolFromNumberParser = "positive_negative"

	// BoolFromNumberParserSignOfOne is the "-1/1" parser.
	// -1 means false, 1 means true. Other numbers are considerd "non parsed" (fallback value or Red result).
	BoolFromNumberSignOfOne BoolFromNumberParser = "sign_of_one"
)

// Config holds global settings for dirty unmarshalling.
type Config struct {
	// Bool is the configuration for dirty.Bool.
	Bool struct {
		FromStrings struct {
			// FromStrings.Allowed allows boolean to be decoded from a string.
			// By default it will only decode bools from "true" and "false" strings.
			//
			// Default: true
			Allowed bool

			// CustomListForTrue specifies list of string values that are considered true.
			// It's ignored if FromStrings.Allowed is false.
			// Values here are case-insensitive.
			//
			// Default: ["true"]
			// Example: ["true", "yes", "on"]
			CustomListForTrue []string

			// CustomListForFalse specifies list of string values that are considered false.
			// It's ignored if FromStrings.Allowed is false.
			// Values here are case-insensitive.
			//
			// Default: ["false"]
			// Example: ["false", "no", "off"]
			CustomListForFalse []string

			// FalseForEmptryString specifies that "" should be considered as false
			// This config option is actually a shortcut for adding a `""` in the CustomListForFalse
			//
			// Default: true
			FalseForEmptyString bool

			// RespectFromNumbersLogic allows to parse stringified number value
			// as a regular number values (corresponding to the FromNumbers config)
			RespectFromNumbersLogic bool

			// FallbackValue is the bool result for string values
			// After not falling into one of the CustomListForTrue/CustomListForFalse lists.
			//
			// Default: options.Some(false) // considered as real false value
			// Example: option.Some(true) 	// will default to true when unmarshalled value
			// 			option.None() 		// can cause red result when unmarshalled value
			FallbackValue option.Bool
		}

		FromNumbers struct {
			// Allowed allows boolean to be decoded from an integer.
			// By default it will only decode bools from 1 or 0 numbers.
			//
			// Default: true
			Allowed bool

			// CustomParseFunc specifies how to parse numbers to bool.
			// Is ignored if FromNumbers.Allowed is false.
			//
			// Default: BoolFromNumberParserBinary (1 is true, 0 is false)
			CustomParseFunc BoolFromNumberParser

			// FallbackValue is the bool result for number values
			// After resulting in option.None result from CustomParseFunc.
			//
			// Default: options.Some(false) // considered as real false value
			// Example: option.Some(true) 	// will default to true when unmarshalled value
			// 			option.None() 		// can cause red result when unmarshalled value
			FallbackValue option.Bool
		}

		FromNull struct {
			// Allowed allows boolean to be decoded from a null.
			// By default it will decode null as false.
			//
			// Default: true
			Allowed bool

			// Inverse means inversing the FromNull logic.
			// If inverse:true nulls will be considered `true` rather than `false` as by default.
			//
			// Default: false
			Inverse bool
		}
	}

	// Number is the configuration for dirty.Number.
	Number struct {
		FromStrings struct {
			// FromStrings.Allowed indicates whether numeric values provided as strings should be accepted.
			//
			// Default: true.
			Allowed bool

			// SpacingAllowed indicates whether the spacing should be trimed in the stringified numbers.
			// Example: "1 000 000" is considered as a valid 1000000 in this case.
			//
			// Default: true.
			SpacingAllowed bool

			// ExponentNotationAllowed specifies whether numeric values with exponent should be accepted.
			// Example: "1e6" is considered as a valid 1000000 in this case.
			//
			// Default: true.
			ExponentNotationAllowed bool

			// CommasAllowed indicates whether numeric values with comma should be accepted.
			// Example: "1,000,000" is considered as a valid 1000000 in this case.
			//
			// Default is true.
			CommasAllowed bool

			// FloatishAllowed indicates whether 1.0 is considered a valid integer accepted in
			// integer-based type in the clean mode.
			// Note: this means that having `V int64 `json:"v"` in your clean (strict) model,
			//       and `V dirty.Number `json:"v"` in your dirty model,
			//       it will successfully forgive the  5.0 for 5 (resulting as Yellow),
			//       but will end up Red and lose the value in case of 5.1.
			//
			// Default is true.
			FloatishAllowed bool
		}

		FromBools struct {
			// Allowed allows number to be decoded from a Bool.
			// By default true is decoded as 1.0 and false as 0.0
			//
			// Default: true
			Allowed bool

			// TODO: maybe custom logic config is needed here?
		}

		FromNull struct {
			// Allowed allows number to be decoded from a null.
			// By default it will decode number as zero.
			//
			// Default: true
			Allowed bool
		}
	}
}

// defaultConfig is the source-of-truth for the default configuration.
func defaultConfig() *Config {
	var cfg Config

	cfg.Bool.FromStrings.Allowed = true
	cfg.Bool.FromStrings.CustomListForTrue = []string{"true", "yes", "on", "1"}
	cfg.Bool.FromStrings.CustomListForFalse = []string{"false", "no", "off", "0"}
	cfg.Bool.FromStrings.FalseForEmptyString = true
	cfg.Bool.FromStrings.FallbackValue = option.Some(false)
	cfg.Bool.FromStrings.RespectFromNumbersLogic = true

	cfg.Bool.FromNumbers.Allowed = true
	cfg.Bool.FromNumbers.CustomParseFunc = BoolFromNumberBinary
	cfg.Bool.FromNumbers.FallbackValue = option.Some(false)

	cfg.Bool.FromNull.Allowed = true
	cfg.Bool.FromNull.Inverse = false

	cfg.Number.FromStrings.Allowed = true
	cfg.Number.FromStrings.SpacingAllowed = true
	cfg.Number.FromStrings.ExponentNotationAllowed = true
	cfg.Number.FromStrings.CommasAllowed = true
	cfg.Number.FromStrings.FloatishAllowed = true
	cfg.Number.FromBools.Allowed = true
	cfg.Number.FromNull.Allowed = true

	return &cfg
}

// globalConfig is the package-level variable storing the config.
var (
	globalConfig *Config
	mu           sync.RWMutex
)

func init() {
	globalConfig = defaultConfig()
}

// Global returns a copy of the global configuration.
// Returned copy is a clone. It's modifying doesn't affect the original config.
func Global() *Config {
	mu.RLock()
	defer mu.RUnlock()

	return clone(globalConfig)
}

// Set updates the global configuration.
// It's often a good idea to validate new values before setting them.
func UpdateGlobal(updateFn func(config *Config)) *Config {
	mu.Lock()
	updateFn(globalConfig)
	defer mu.Unlock()

	return clone(globalConfig)
}

// clone via json round-trip. It's a simple (but not the most efficient) way to clone the config.
// Config is safe for marshalling (that's by design): It will never contain functions, etc.
// We can live with this solution until we need increase performance.
func clone(cfg *Config) *Config {
	marshalled, _ := json.Marshal(cfg)
	var cloned Config
	_ = json.Unmarshal(marshalled, &cloned)
	return &cloned
}
