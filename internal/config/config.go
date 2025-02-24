package config

import (
	"sync"
)

// BoolConfig defines how dirty the dirty.Bool is decoded.
type BoolConfig struct {
	// AllowString allows boolean to be decoded from a string.
	// By default it only supports "true" and "false".
	AllowString bool // allows "true" and "false"

	// TrueStrings specifies list of string values that are considered true.
	// When no specified, but AllowString:true, the TrueStrings is considered as ["true"].
	// Note: values here are case-insensitive.
	TrueStrings []string // e.g. ["true", "yes", "on"]

	// FalseStrings specifies list of string values that are considered false.
	// When no specified, but AllowString:true, the FalseStrings is considered as ["false"].
	// Note: values here are case-insensitive.
	FalseStrings []string // e.g. ["false", "no", "off"]

	// FallbackStringValue is the bool result for string values not specified in TrueStrings or FalseStrings.
	// By default it's false.
	FallbackStringValue bool

	// FallbackFromStringToRed makes the result "Red" for string values not specified in TrueStrings or FalseStrings.
	// This means the current bool is undefined, and we lose it.
	FallbackFromStringToRed bool

	// AllowNumber allows boolean to be decoded from an integer.
	// By default it only supports 1 and 0.
	AllowNumber bool // allows 1 and 0

	// TrueNumbers specifies list of number values that are considered true.
	// When no specified, but AllowNumber:true, the TrueNumbers is considered as func(f float64) bool{f == 1}.
	// Note: values here can be both integers and floats. 1.0 and 1 are treated as the same.
	TrueNumbers func(float64) bool // e.g. func (f float64) bool { return f > 0 }

	// FalseNumbers specifies list of number values that are considered false.
	// When no specified, but AllowNumber:true, the FalseNumbers is considered as func(f float64) bool{f == 0}.
	// Note: values here can be both integers and floats. 1.0 and 1 are treated as the same.
	FalseNumbers func(float64) bool // e.g. func (f float64) bool { return f <= 0 }

	// FallbackNumberValue is the bool result for number values not specified in TrueNumbers or FalseNumbers..
	// By default it's false.
	// Note: FallbackNumberValue's behavior depends on how you declare TrueNumbers and FalseNumbers.
	//       If TrueNumbers/FalseNumbers are covering all numbers (so one of them is always true), so fallback never happens.
	FallbackNumberValue bool

	// FallbackFromNumberToRed makes the result "Red" for number values not specified in TrueNumbers or FalseNumbers.
	// This means the current bool is undefined, and we lose it.
	FallbackFromNumberToRed bool
}

// NumberConfig defines how dirty numbers are decoded.
type NumberConfig struct {
	// AllowString indicates whether numeric values provided as strings should be accepted.
	// Default is true.
	AllowString bool

	// AllowSpacing indicates whether numeric values with spacing should be accepted.
	// Note: "1 000 000" is considered as a valid 1000000 in this case.
	// Default is true.
	AllowSpacing bool

	// AllowExponent indicates whether numeric values with exponent should be accepted.
	// Note: "1e6" is considered as a valid 1000000 in this case.
	// Default is true.
	AllowExponent bool

	// AllowComma indicates whether numeric values with comma should be accepted.
	// Note: "1,000,000" is considered as a valid 1000000 in this case.
	// Default is true.
	AllowComma bool

	// AllowFloatishIntegers indicates whether 1.0 is considered a valid integer accepted in
	// integer-based type in the clean mode.
	// Note: this means that having `V int64 `json:"v"` in your clean (strict) model,
	//       and `V dirty.Number `json:"v"` in your dirty model,
	//       it will successfully forgive the  5.0 for 5 (resulting as Yellow),
	//       but will end up Red and lose the value in case of 5.1.
	// Default is true.
	AllowFloatishIntegers bool
}

// Config holds global settings for dirty unmarshalling.
type Config struct {
	// Bool is the configuration for dirty.Bool.
	Bool BoolConfig

	// Number is the configuration for dirty.Number.
	Number NumberConfig
}

// globalConfig is the package-level variable storing the config.
var (
	globalConfig *Config
	once         sync.Once
	mu           sync.RWMutex
)

func initConfig() {
	globalConfig = &Config{
		Bool: BoolConfig{
			AllowString:             true,
			TrueStrings:             []string{"true", "yes", "on", "1"},
			FalseStrings:            []string{"false", "no", "off", "0"},
			FallbackStringValue:     false,
			FallbackFromStringToRed: false,
			AllowNumber:             true,
			TrueNumbers:             func(f float64) bool { return f == 0 },
			FalseNumbers:            func(f float64) bool { return f != 0 },
			FallbackNumberValue:     false,
			FallbackFromNumberToRed: false,
		},
		Number: NumberConfig{
			AllowString:           true,
			AllowSpacing:          true,
			AllowExponent:         true,
			AllowComma:            true,
			AllowFloatishIntegers: true,
		},
	}
}

// Get returns a copy of the global configuration.
func Get() Config {
	once.Do(initConfig) // ensure it's initialized only once
	mu.RLock()
	defer mu.RUnlock()
	// Return a copy to avoid race conditions if the caller modifies it.
	return *globalConfig
}

// Set updates the global configuration.
// It's often a good idea to validate new values before setting them.
func Set(newConfig Config) {
	once.Do(initConfig) // ensure it's initialized if not already
	mu.Lock()
	defer mu.Unlock()
	*globalConfig = newConfig
}
