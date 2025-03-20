package config

import (
	"embed"
	"io/fs"

	"github.com/d3rty/json/internal/option"
	"github.com/pelletier/go-toml/v2"
)

//go:embed default.toml
var embeddedConfig embed.FS

// TODO: FromNull behavior should be done via Option
// So, if dirty model has the Option type, then FromNull should respect the option type.

// TODO: allow read single json into array (so just first item is filled)
// and probably opposite (showing in red how much data was lost, but first was set)

// Config holds global settings for dirty unmarshalling.
type Config struct {
	Bool struct {
		Allowed bool

		FallbackValue option.Bool

		FromStrings struct {
			Allowed                 bool
			CustomListForTrue       []string
			CustomListForFalse      []string
			CaseInsensitive         bool
			FalseForEmptyString     bool
			RespectFromNumbersLogic bool
		}
		FromNumbers struct {
			Allowed         bool
			CustomParseFunc BoolFromNumberAlg
		}
		FromNull struct {
			Allowed bool
			Inverse bool
		}
	}
	Number struct {
		Allowed     bool
		FromStrings struct {
			Allowed                 bool
			SpacingAllowed          bool
			ExponentNotationAllowed bool
			CommasAllowed           bool
			RoundingAlgorithm       RoundingAlg
		}
		FromBools struct {
			Allowed bool
		}
		FromNull struct {
			Allowed bool
		}
	}
	Date struct {
		Allowed  bool
		Timezone struct {
			Default             string
			Fields              []string
			ForceConvertingInto bool
		}
		FromNumbers struct {
			Allowed            bool
			UnixTimestamp      bool
			UnixMilliTimestamp bool
		}
		FromStrings struct {
			Allowed bool
			Layouts struct {
				Time     []string
				Date     []string
				DateTime []string
			}
			Aliases                 []string
			RespectFromNumbersLogic bool
		}
		FromNull struct {
			Allowed bool
		}
	}
	FlexKeys struct {
		Allowed         bool
		CaseInsensitive bool
		ChameleonCase   bool
	}
}

// FromBytes read config from given raw []byte.
func FromBytes(data []byte) *Config {
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	return &cfg
}

// String shows string represenatation of the config. It used primarily for debug purposes or verbose mode
// We use `toml` representation here.
func (cfg *Config) String() string {
	j, _ := toml.Marshal(cfg)
	return string(j)
}

// TODO precache in variable.
func defaultConfig() *Config {
	data, err := fs.ReadFile(embeddedConfig, "default.toml")
	if err != nil {
		panic("failed to read embedded default config " + err.Error())
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		panic("failed to unmarshal default.toml config: " + err.Error())
	}

	return &cfg
}

// cleanConfig returns config that disables all dirty options
// unmarshalling with clean config behaves the same as clean starndard unmarshalling.
func cleanConfig() *Config {
	var cfg Config
	return &cfg
}

// ResetToEmpty resets config to its clean state (clean config).
func (cfg *Config) ResetToEmpty() { *cfg = *cleanConfig() }

// ResetToDefault resets config to the default state.
func (cfg *Config) ResetToDefault() { *cfg = *defaultConfig() }

// clone via toml round-trip. It's a simple (but not the most efficient) way to clone the config.
// Config is safe for marshalling (that's by design): It will never contain functions, etc.
// We can live with this solution until we need increase performance.
// TODO: remove panics for live prod code.
func clone(cfg *Config) *Config {
	contents, err := toml.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	var clone Config
	if err := toml.Unmarshal(contents, &clone); err != nil {
		panic(err)
	}
	return &clone
}
