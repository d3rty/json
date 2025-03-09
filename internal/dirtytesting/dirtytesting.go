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

type DirtifyCfg struct {
	rng      *rand.Rand
	cfg      *config.Config
	ratio    float64
	allowRed bool
}

type drtfOpt func(*DirtifyCfg)

func (dcfg *DirtifyCfg) Config() *config.Config { return dcfg.cfg }

func WithConfig(cfg *config.Config) drtfOpt { return func(dcfg *DirtifyCfg) { dcfg.cfg = cfg } }
func WithRng(rng *rand.Rand) drtfOpt        { return func(dcfg *DirtifyCfg) { dcfg.rng = rng } }
func WithRatio(r float64) drtfOpt           { return func(dcfg *DirtifyCfg) { dcfg.ratio = r } }
func WithAllowedRed(b bool) drtfOpt         { return func(dcfg *DirtifyCfg) { dcfg.allowRed = b } }

// Dirtify makes a dirty version of JSON
func Dirtify[T any](cleanJSON []byte, dcfg *DirtifyCfg, opts ...drtfOpt) ([]byte, error) {
	if dcfg == nil {
		if len(opts) == 0 {
			panic("def something wrong. if using default random, you must know it back. pass empty dcfg then")
		}

		dcfg = &DirtifyCfg{}
	}
	dcfg.rng = newRng()
	dcfg.cfg = RandomConfig(dcfg.rng)
	dcfg.ratio = 0.7

	// override dirtify config
	for _, opt := range opts {
		opt(dcfg)
	}

	// Unmarshal clean JSON into the provided model.
	var cleanModel T
	if err := json.Unmarshal(cleanJSON, &cleanModel); err != nil {
		return nil, fmt.Errorf("failed to unmarshal clean JSON: %w", err)
	}

	dirtyModel := NewDirtyfier(dcfg.ratio, dcfg.cfg, dcfg.rng).Dirtify(
		structToMap(cleanModel),
	)

	// Marshal back to JSON.
	dirtyJSON, err := json.Marshal(dirtyModel)
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
