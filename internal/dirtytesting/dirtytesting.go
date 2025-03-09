package dirtytesting

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"slices"
	"time"

	"github.com/d3rty/json/internal/config"
	"github.com/d3rty/json/internal/option"
)

func newRng() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// RandomConfig returns a randomly generated *Config.
func RandomConfig(r *rand.Rand) *config.Config {
	if r == nil {
		r = newRng()
	}

	cfg := new(config.Config)

	// --- Bool Configuration ---
	// Randomly decide if dirty bool is allowed.
	cfg.Bool.Allowed = r.Intn(2) == 0
	if cfg.Bool.Allowed {
		// FromStrings
		cfg.Bool.FromStrings.Allowed = r.Intn(2) == 0
		if cfg.Bool.FromStrings.Allowed {
			// Generate a random preset for "true" values (between 4 and 6 values)
			cfg.Bool.FromStrings.CustomListForTrue = generateRandomPreset(dictTrues, 4, 6, r)
			// Generate a random preset for "false" values (between 4 and 6 values)
			cfg.Bool.FromStrings.CustomListForFalse = generateRandomPreset(dictFalses, 4, 6, r)
			cfg.Bool.FromStrings.FalseForEmptyString = r.Intn(2) == 0
			cfg.Bool.FromStrings.RespectFromNumbersLogic = r.Intn(2) == 0
			// Fallback value as a random boolean.
			cfg.Bool.FromStrings.FallbackValue = option.Some(r.Intn(2) == 0)
		}

		// FromNumbers
		cfg.Bool.FromNumbers.Allowed = r.Intn(2) == 0
		if cfg.Bool.FromNumbers.Allowed {
			choices := config.AvailableBoolFromNumberParsers()
			cfg.Bool.FromNumbers.CustomParseFunc = choices[r.Intn(len(choices))]
			cfg.Bool.FromNumbers.FallbackValue = option.Some(r.Intn(2) == 0)
		}

		// FromNull
		cfg.Bool.FromNull.Allowed = r.Intn(2) == 0
		if cfg.Bool.FromNull.Allowed {
			cfg.Bool.FromNull.Inverse = r.Intn(2) == 0
		}
	}

	// --- Number Configuration ---
	cfg.Number.Allowed = r.Intn(2) == 0
	if cfg.Number.Allowed {
		// FromStrings
		cfg.Number.FromStrings.Allowed = r.Intn(2) == 0
		if cfg.Number.FromStrings.Allowed {
			cfg.Number.FromStrings.SpacingAllowed = r.Intn(2) == 0
			cfg.Number.FromStrings.ExponentNotationAllowed = r.Intn(2) == 0
			cfg.Number.FromStrings.CommasAllowed = r.Intn(2) == 0
			cfg.Number.FromStrings.FloatishAllowed = r.Intn(2) == 0
		}

		// FromBools
		cfg.Number.FromBools.Allowed = r.Intn(2) == 0

		// FromNull
		cfg.Number.FromNull.Allowed = r.Intn(2) == 0
	}

	// --- FlexKeys Configuration ---
	cfg.FlexKeys.Allowed = r.Intn(2) == 0
	if cfg.FlexKeys.Allowed {
		cfg.FlexKeys.CaseInsensitive = r.Intn(2) == 0
		cfg.FlexKeys.ChameleonCase = r.Intn(2) == 0
	}

	return cfg
}

func GenerateDirtyJSON(model any, cleanJSON []byte, ratio float64, allowRedArg ...bool) ([]byte, error) {
	// Allow red (lossy transformations) if requested.
	var allowRed bool
	if len(allowRedArg) > 0 {
		allowRed = allowRedArg[0]
	}
	if allowRed {
		panic("not implemented")
	}

	// Unmarshal clean JSON into the provided model.
	if err := json.Unmarshal(cleanJSON, model); err != nil {
		return nil, fmt.Errorf("failed to unmarshal clean JSON: %w", err)
	}

	rng := newRng()

	mixedData := NewDirtyfier(ratio, RandomConfig(rng), rng).Dirtify(structToMap(model))

	// Marshal back to JSON.
	dirtyJSON, err := json.Marshal(mixedData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dirty JSON: %w", err)
	}
	return dirtyJSON, nil
}

var dictTrues = []string{"true", "yes", "on", "1", "ok", "yep"}
var dictFalses = []string{"false", "no", "off", "0", "nah", "nope"}

// generateRandomPreset selects a random subset (of size between min and max)
// from the provided master list.
func generateRandomPreset(dict []string, min, max int, r *rand.Rand) []string {
	// Determine the number of elements to pick.
	count := r.Intn(max-min+1) + min
	if len(dict) < count {
		count = len(dict)
	}

	// Shuffle the master copy.
	shuffled := slices.Clone(dict)
	r.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Return the first count elements.
	return shuffled[:count]
}
