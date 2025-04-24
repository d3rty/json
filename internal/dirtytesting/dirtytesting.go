package dirtytesting

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/d3rty/json/internal/config"
	"github.com/d3rty/json/internal/flipping"
	"github.com/d3rty/json/internal/option"
)

// RandomConfig returns a randomly generated *Config.
func RandomConfig(coinArg ...*flipping.Coin) *config.Config {
	coin := flipping.MaybeNewCoin(coinArg...)

	cfg := config.New()

	// --- Bool Configuration ---
	// Randomly decide if dirty bool is allowed.
	if cfg.Bool.Disabled = coin.Flip(); !cfg.Bool.Disabled {
		cfg.Bool.FallbackValue = option.Some(coin.Flip())

		// Bool.FromStrings
		if cfg.Bool.FromStrings.Disabled = coin.Flip(); cfg.Bool.FromStrings.Disabled {
			dictMinSize, dictMaxSize := 3, 6
			// Generate a random preset for "true" values (between 3 and 6 values)
			cfg.Bool.FromStrings.CustomListForTrue = generateRandomPreset(dictTrues, dictMinSize, dictMaxSize, coin)
			// Generate a random preset for "false" values (between 3 and 6 values)
			cfg.Bool.FromStrings.CustomListForFalse = generateRandomPreset(dictFalses, dictMinSize, dictMaxSize, coin)
			cfg.Bool.FromStrings.RespectFromNumbersLogic = coin.Flip()
			// Fallback value as a random boolean.
		}

		// Bool.FromNumbers
		if cfg.Bool.FromNumbers.Disabled = coin.Flip(); !cfg.Bool.FromNumbers.Disabled {
			cfg.Bool.FromNumbers.CustomParseFunc = flipping.FeelingLucky(
				config.ListAvailableBoolFromNumberAlgs(),
				coin,
			)
		}

		// FromNull
		if cfg.Bool.FromNull.Disabled = coin.Flip(); cfg.Bool.FromNull.Disabled {
			cfg.Bool.FromNull.Inverse = coin.Flip()
		}
	}

	// --- Number Configuration ---
	if cfg.Number.Disabled = coin.Flip(); !cfg.Number.Disabled {
		// FromStrings
		if cfg.Number.FromStrings.Disabled = coin.Flip(); !cfg.Number.FromStrings.Disabled {
			cfg.Number.FromStrings.SpacingAllowed = coin.Flip()
			cfg.Number.FromStrings.ExponentNotationAllowed = coin.Flip()
			cfg.Number.FromStrings.CommasAllowed = coin.Flip()

			cfg.Number.FromStrings.RoundingAlgorithm = flipping.FeelingLucky(
				config.ListAvailableRoundingAlgs(),
				coin,
			)
		}

		// FromBools
		cfg.Number.FromBools.Disabled = coin.Flip()

		// FromNull
		cfg.Number.FromNull.Disabled = coin.Flip()
	}

	// --- FlexKeys Configuration ---
	if cfg.FlexKeys.Disabled = coin.Flip(); !cfg.FlexKeys.Disabled {
		cfg.FlexKeys.CaseInsensitive = coin.Flip()
		cfg.FlexKeys.ChameleonCase = coin.Flip()
	}

	return cfg
}

type DirtifyCfg struct {
	coin     *flipping.Coin
	cfg      *config.Config
	ratio    float64
	allowRed bool
}

type Opt func(*DirtifyCfg)

func (dcfg *DirtifyCfg) Config() *config.Config { return dcfg.cfg }

func WithConfig(cfg *config.Config) Opt { return func(dcfg *DirtifyCfg) { dcfg.cfg = cfg } }
func WithCoin(coin *flipping.Coin) Opt  { return func(dcfg *DirtifyCfg) { dcfg.coin = coin } }
func WithRatio(r float64) Opt           { return func(dcfg *DirtifyCfg) { dcfg.ratio = r } }
func WithAllowedRed(b bool) Opt         { return func(dcfg *DirtifyCfg) { dcfg.allowRed = b } }

// Dirtify makes a dirty version of JSON.
func Dirtify[T any](cleanJSON []byte, dcfg *DirtifyCfg, opts ...Opt) ([]byte, error) {
	if dcfg == nil {
		if len(opts) == 0 {
			panic("def something wrong. if using default random, you must know it back. pass empty dcfg then")
		}

		dcfg = &DirtifyCfg{}
	}

	if len(opts) == 0 {
		dcfg.coin = flipping.NewCoin()
		dcfg.cfg = RandomConfig(dcfg.coin)
		dcfg.ratio = 0.7
	} else {
		// override dirtify config
		for _, opt := range opts {
			opt(dcfg)
		}
	}

	// Unmarshal clean JSON into the provided model.
	var cleanModel T
	if err := json.Unmarshal(cleanJSON, &cleanModel); err != nil {
		return nil, fmt.Errorf("failed to unmarshal clean JSON: %w", err)
	}

	dirtyModel := NewDirtifier(dcfg.ratio, dcfg.cfg, dcfg.coin).Make(
		StructToMap(cleanModel),
	)

	// Marshal back to JSON.
	dirtyJSON, err := json.Marshal(dirtyModel)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dirty JSON: %w", err)
	}
	return dirtyJSON, nil
}

// dictTrues is a dictionary for string values of True
//
//nolint:gochecknoglobals // because we can
var dictTrues = []string{"true", "yes", "on", "1", "ok", "yep"}

// dictFalses is a dictionary for string values of False
//
//nolint:gochecknoglobals // because we can
var dictFalses = []string{"false", "no", "off", "0", "nah", "nope", ""}

// generateRandomPreset selects a random subset (of size between min and max)
// from the provided dict list.
func generateRandomPreset(dict []string, from, to int, coinArg ...*flipping.Coin) []string {
	coin := flipping.MaybeNewCoin(coinArg...)

	// Determine the number of elements to pick.
	count := min(
		coin.Rng().Intn(to-from+1)+from,
		len(dict),
	)

	// Shuffle the dict copy.
	shuffled := slices.Clone(dict)
	coin.Rng().Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Return the first count elements.
	return shuffled[:count]
}
