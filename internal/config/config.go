package config

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/d3rty/json/internal/option"
)

//go:embed default.toml
var embeddedConfig embed.FS

// Section is a config section (toml table).
// Section is considered disabled if Disabled=true or full its parent is nil (see cfg.init()).
type Section struct {
	Disabled bool `toml:"Disabled"`
}

func (s Section) IsDisabled() bool { return s.Disabled }

// Config holds global settings for dirty unmarshalling.
type Config struct {
	Section

	Bool     *BoolConfig     `toml:"Bool"`
	Number   *NumberConfig   `toml:"Number"`
	Date     *DateConfig     `toml:"Date"`
	Array    *ArrayConfig    `toml:"Array"`
	FlexKeys *FlexKeysConfig `toml:"FlexKeys"`
}

type BoolConfig struct {
	Section

	FallbackValue option.Bool `toml:"FallbackValue"`

	FromStrings *BoolFromStringsConfig `toml:"FromStrings"`
	FromNumbers *BoolFromNumbersConfig `toml:"FromNumbers"`
	FromNull    *BoolFromNullConfig    `toml:"FromNull"`
}

type (
	BoolFromStringsConfig struct {
		Section

		// TODO: precache customListForTrue + CaseInsensitive (so we don't do lower of the list each time)

		CustomListForTrue       []string `toml:"CustomListForTrue"`
		CustomListForFalse      []string `toml:"CustomListForFalse"`
		CaseInsensitive         bool     `toml:"CaseInsensitive"`
		RespectFromNumbersLogic bool     `toml:"RespectFromNumbersLogic"`
	}

	BoolFromNumbersConfig struct {
		Section

		CustomParseFunc BoolFromNumberAlg `toml:"CustomParseFunc"`
	}

	BoolFromNullConfig struct {
		Section

		Inverse bool `toml:"Inverse"`
	}
)

type NumberConfig struct {
	Section

	FromStrings *NumberFromStringsConfig `toml:"FromStrings"`
	FromBools   *NumberFromBoolsConfig   `toml:"FromBools"`
	FromNull    *NumberFromNullConfig    `toml:"FromNull"`
}

type (
	NumberFromStringsConfig struct {
		Section

		SpacingAllowed          bool        `toml:"SpacingAllowed"`
		ExponentNotationAllowed bool        `toml:"ExponentNotationAllowed"`
		CommasAllowed           bool        `toml:"CommasAllowed"`
		RoundingAlgorithm       RoundingAlg `toml:"RoundingAlgorithm"`
	}
	NumberFromBoolsConfig struct {
		Section
	}
	NumberFromNullConfig struct {
		Section
	}
)

type DateConfig struct {
	Section

	Timezone    *DateTimezoneConfig    `toml:"Timezone"`
	FromNumbers *DateFromNumbersConfig `toml:"FromNumbers"`
	FromStrings *DateFromStringsConfig `toml:"FromStrings"`
	FromNull    *DateFromNullConfig    `toml:"FromNull"`
}

type (
	DateTimezoneConfig struct {
		Section

		Default             string   `toml:"Default"`
		Fields              []string `toml:"Fields"`
		ForceConvertingInto bool     `toml:"ForceConvertingInto"`
	}
	DateFromNumbersConfig struct {
		Section

		UnixTimestamp      bool `toml:"UnixTimestamp"`
		UnixMilliTimestamp bool `toml:"UnixMilliTimestamp"`
	}
	DateFromStringsConfig struct {
		Section

		Layouts struct {
			Time     []string `toml:"Time"`
			Date     []string `toml:"Date"`
			DateTime []string `toml:"DateTime"`
		}
		Aliases                 bool `toml:"Aliases"`
		RespectFromNumbersLogic bool `toml:"RespectFromNumbersLogic"`
	}
	DateFromNullConfig struct {
		Section
	}
)

type ArrayConfig struct {
	Section

	// AutoWrapSingleValues allows to have `result: x` to be considered as `result: [x]`
	AutoWrapSingleValues bool `toml:"AutoWrapSingleValues"`
}

type FlexKeysConfig struct {
	Section

	CaseInsensitive bool `toml:"CaseInsensitive"`
	ChameleonCase   bool `toml:"ChameleonCase"`
}

// FromBytes read config from a given raw [] byte.
func FromBytes(data []byte) *Config {
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	return (&cfg).init()
}

// String shows string representation of the config. It used primarily for debug purposes or verbose mode
// We use `toml` representation here.
func (cfg *Config) String() string {
	j, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Sprintf("<<invalid config>>\n%s", err)
	}

	return string(j)
}

// Clone returns the deep safe copy of a Config instance.
func (cfg *Config) Clone() *Config {
	cloned, err := clone(cfg)
	if err != nil {
		panic(err)
	}

	return cloned
}

// init sets default for all the config fields.
// if a subconfig is nil, it automatically changes it to an empty config with Disabled=true.
func (cfg *Config) init() *Config {
	handleDefaultFieldDisabled(cfg)

	if cfg.Date.Timezone.Default == "" {
		cfg.Date.Timezone.Default = "UTC"
	}
	if len(cfg.Date.Timezone.Fields) == 0 {
		cfg.Date.Timezone.Fields = []string{"tz", "timezone"}
	}

	return cfg
}

// newConfig returns a new (empty/clean) config that disables all dirty options.
// dirty unmarshalling with an empty config behaves the same as standard unmarshalling.
func newConfig() *Config {
	return (&Config{}).init()
}

// ResetToEmpty resets config to its clean state (clean config).
func (cfg *Config) ResetToEmpty() { *cfg = *newConfig() }

// ResetToDefault resets config to the default state.
func (cfg *Config) ResetToDefault() { *cfg = *Default() }

// New creates a new and empty config.
func New() *Config { return newConfig() }

func Load(path string) (*Config, error) {
	// #nosec G304 -- we're OK with loading dynamic variable
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		panic("failed to unmarshal default.toml config: " + err.Error())
	}

	return cfg.init(), nil
}

var (
	defaultCfg  *Config
	defaultOnce sync.Once
)

// Default returns a copy of the loaded default config (default.toml).
func Default() *Config {
	defaultOnce.Do(func() {
		data, err := fs.ReadFile(embeddedConfig, "default.toml")
		if err != nil {
			panic("failed to read embedded default config " + err.Error())
		}

		var cfg Config
		if err := toml.Unmarshal(data, &cfg); err != nil {
			panic("failed to unmarshal default.toml config: " + err.Error())
		}

		defaultCfg = (&cfg).init()
	})

	// Return a deep copy to prevent modification of the cached config
	return defaultCfg.Clone()
}
