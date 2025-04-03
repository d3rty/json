package config

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/d3rty/json/internal/option"
)

//go:embed default.toml
var embeddedConfig embed.FS

// TODO: FromNull behavior should be done via Option
// So, if dirty model has the Option type, then FromNull should respect the option type.

// TODO: allow read single json into array (so just first item is filled)
// and probably opposite (showing in red how much data was lost, but first was set)

// Section is a config section (toml table).
// Section is considered disabled if Disabled=true or full its parent is nil (see cfg.Init()).
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

type FlexKeysConfig struct {
	Section

	CaseInsensitive bool `toml:"CaseInsensitive"`
	ChameleonCase   bool `toml:"ChameleonCase"`
}

// FromBytes read config from given raw []byte.
func FromBytes(data []byte) *Config {
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	return (&cfg).Init()
}

// String shows string represenatation of the config. It used primarily for debug purposes or verbose mode
// We use `toml` representation here.
func (cfg *Config) String() string {
	j, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Sprintf("<<invalid config>>\n%s", err)
	}

	return string(j)
}

// Init sets default for all the config fields.
// If a subconfig is nil, it automatically changes it to empty config with Disabled=true.
func (cfg *Config) Init() *Config {
	handleDefaultFieldDisabled(cfg)

	if cfg.Date.Timezone.Default == "" {
		cfg.Date.Timezone.Default = "UTC"
	}
	if len(cfg.Date.Timezone.Fields) == 0 {
		cfg.Date.Timezone.Fields = []string{"tz", "timezone"}
	}

	return cfg
}

// defaultConfigs returns a copy of the default config.
func defaultConfig() *Config {
	data, err := fs.ReadFile(embeddedConfig, "default.toml")
	if err != nil {
		panic("failed to read embedded default config " + err.Error())
	}

	// TODO precache in variable.

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		panic("failed to unmarshal default.toml config: " + err.Error())
	}

	cfg.Init()

	return &cfg
}

// newConfig returns a new (empty/clean) config that disables all dirty options.
// dirty unmarshalling with an empty config behaves the same as starndard unmarshalling.
func newConfig() *Config {
	return (&Config{}).Init()
}

// ResetToEmpty resets config to its clean state (clean config).
func (cfg *Config) ResetToEmpty() { *cfg = *newConfig() }

// ResetToDefault resets config to the default state.
func (cfg *Config) ResetToDefault() { *cfg = *defaultConfig() }

// New creates a new and empty config.
func New() *Config { return newConfig() }

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		panic("failed to unmarshal default.toml config: " + err.Error())
	}

	return cfg.Init(), nil
}
