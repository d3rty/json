package dirtytesting

import (
	"fmt"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/d3rty/json/cases"
	"github.com/d3rty/json/internal/config"
	"github.com/d3rty/json/internal/flipping"
)

// Dirtifier makes dirty JSONs from clean ones.
type Dirtifier struct {
	// threshold is a number from 0.0 to 1.0 that sets how much "dirty" the result will end up.
	// e.g., 1.0 means that 100% of possible fields will be dirtified.
	//      0.0 means that 0% of possible fields will be dirtified (so the result will remain clean).
	threshold float64

	// coin is a wrapper around the random generator
	// it allows to make random decision for each field to be dirtified
	coin *flipping.Coin

	// cfg stands for dirty Config: how specifically we do dirty.
	cfg *config.Config
}

// NewDirtifier creates a new dirtifier.
func NewDirtifier(threshold float64, cfg *config.Config, coinArg ...*flipping.Coin) *Dirtifier {
	return &Dirtifier{
		coin:      flipping.MaybeNewCoin(coinArg...),
		cfg:       cfg,
		threshold: threshold,
	}
}

// keepItClean returns true/false depending on the current rand generation and the threshold
// keepItClean returns true if we should omit dirtifying, and false if we should do dirtifying.
func (d *Dirtifier) keepItClean() bool { return d.coin.Chance(d.threshold) }

func (d *Dirtifier) randomCase(s string) string {
	// Optionally, seed the random number generator once in your main function or init block.
	// rand.Seed(time.Now().UnixNano())

	// Convert the string to a slice of runes to properly handle Unicode characters.
	runes := []rune(s)
	for i, r := range runes {
		// With a 50% chance convert to lower case, else upper case.
		if d.coin.Flip() {
			runes[i] = unicode.ToLower(r)
		} else {
			runes[i] = unicode.ToUpper(r)
		}
	}
	return string(runes)
}

// Make makes dirty values applying random dirtify-transformations.
func (d *Dirtifier) Make(val any) any {
	switch v := val.(type) {
	case map[string]any:
		// Process map values.
		newMap := make(map[string]any)
		for key, elem := range v {
			newMap[d.makeDirtyKey(key)] = d.Make(elem)
		}
		return newMap
	case []any:
		// Process each element of the slice.
		for i, elem := range v {
			v[i] = d.Make(elem)
		}
		return v
	case bool:
		// With probability, transform booleans to allowed string representations.
		return d.makeDirtyBool(v)
	case float64:
		// With probability, transform numbers to string.
		return d.makeDirtyNumber(v)
	case string:
		// strings are not dirtified yet
		return v
	default:
		// For other types (strings, nil, etc.), leave unchanged.
		return v
	}
}

// makeDirtyKey makes a dirty key from a given clean key.
func (d *Dirtifier) makeDirtyKey(key string) string {
	if d.cfg.FlexKeys.IsDisabled() {
		return key
	}
	if !d.cfg.FlexKeys.CaseInsensitive && !d.cfg.FlexKeys.ChameleonCase {
		return key
	}
	if d.keepItClean() {
		return key
	}

	if cases.IsComplexCase(key) && d.cfg.FlexKeys.ChameleonCase && d.coin.Flip() {
		// TODO: it would be great to exclude current case (so e.g. cases.MatchCase(key))
		allCases := []cases.Case{
			cases.Camel,
			cases.Snake,
			cases.Kebab,
			cases.Pascal,
			cases.Header,
			cases.TitleSnake,
			cases.Hybrid,
		}

		convertTo := flipping.FeelingLucky(allCases, d.coin)
		if convertTo == cases.Hybrid {
			return TransformToHybridCase(key, d.coin)
		}

		return cases.TransformTo(key, convertTo)
	}

	if d.cfg.FlexKeys.CaseInsensitive && d.coin.Flip() {
		// Let's mix the case (make it upper/lower/title/etc)

		// We have to shuffle and try them until we get first transformation that makes sense
		// This is made in case the key is already Upper so ToUpper transform will be ignored.
		transformations := []func(string) string{
			strings.ToUpper,
			strings.ToLower,
			strings.ToTitle,
		}
		rand.Shuffle(len(transformations), func(i, j int) {
			transformations[i], transformations[j] = transformations[j], transformations[i]
		})

		for _, transform := range transformations {
			if transformed := transform(key); transformed != key {
				return transformed
			}
		}

		return key
	}

	return key
}

// makeDirtyBool makes a dirty bool from given clean bool.
func (d *Dirtifier) makeDirtyBool(v bool) any {
	if d.cfg.Bool.IsDisabled() || d.keepItClean() {
		return v
	}

	cfg := d.cfg.Bool

	// Let's extract the available flows from the config
	// FromNull is handled separately, as it's considered an additional setting
	// to FromStrings and/or FromNumbers.
	var flows []string
	if !cfg.FromStrings.IsDisabled() {
		flows = append(flows, "string")
	}
	if !cfg.FromNumbers.IsDisabled() {
		flows = append(flows, "number")
	}
	// If no conversion is allowed, return the clean boolean.
	if len(flows) == 0 {
		return v
	}

	// TODO: support .FallbackValue as it's not so simple

	// Randomly choose one conversion flow.
	var numberToBeStringified bool
	switch flipping.FeelingLucky(flows, d.coin) {
	case "string":
		cfgFromStrings := d.cfg.Bool.FromStrings

		// if we respect numbers logic then with 50% chance generate a stringish number
		// e.g. "1", "0", etc.
		// Here we can fallthrough into the "number" case if we simply flip the true coin,
		// or we have no option (custom strings are disabled in config).
		if cfgFromStrings.RespectFromNumbersLogic {
			customStringsDisabled := !v && len(cfgFromStrings.CustomListForFalse) == 0 ||
				v && len(cfgFromStrings.CustomListForTrue) == 0

			if d.coin.Flip() || customStringsDisabled {
				numberToBeStringified = true
			}
		}

		if !numberToBeStringified {
			if v {
				var sTrue = "true" // by default, the "true" string is considered to be as true
				if len(cfgFromStrings.CustomListForTrue) > 0 {
					sTrue = flipping.FeelingLucky(cfgFromStrings.CustomListForTrue, d.coin)
				}

				if cfgFromStrings.CaseInsensitive && d.coin.Flip() {
					sTrue = d.randomCase(sTrue)
				}
				return d.maybeBoolNilify(v, sTrue)
			}

			var sFalse = "false"
			if len(
				cfgFromStrings.CustomListForFalse,
			) > 0 { // by default, the "false" string is considered to be as false
				sFalse = flipping.FeelingLucky(cfgFromStrings.CustomListForFalse, d.coin)
			}
			if cfgFromStrings.CaseInsensitive && d.coin.Flip() {
				sFalse = d.randomCase(sFalse)
			}
			return d.maybeBoolNilify(v, sFalse)
		}

		// numberToBeStringified = true
		fallthrough
	case "number":
		cfg := d.cfg.Bool.FromNumbers

		var intBoolResult int
		switch cfg.CustomParseFunc {
		case config.BoolFromNumberBinary:
			if v {
				intBoolResult = 1
			} else {
				intBoolResult = 0 // explicitly setting anyway for clarity
			}
		case config.BoolFromNumberPositiveNegative:
			const k = 1000
			if v {
				intBoolResult = d.coin.Rng().Intn(k) + 1
			} else {
				intBoolResult = -d.coin.Rng().Intn(k)
			}
		case config.BoolFromNumberSignOfOne:
			if v {
				intBoolResult = 1
			} else {
				intBoolResult = -1
			}
		case config.BoolFromNumberUndefined:
			fallthrough
		default:
			panic("something really bad")
		}

		if numberToBeStringified {
			return d.maybeBoolNilify(v, strconv.Itoa(intBoolResult))
		}
		return d.maybeBoolNilify(v, intBoolResult)
	default:
		panic("unreachable")
	}
}

// makeDirtyNumber makes a dirty number from a given clean number.
func (d *Dirtifier) makeDirtyNumber(v float64) any {
	if d.cfg.Number.IsDisabled() || d.keepItClean() {
		return v
	}

	cfg := d.cfg.Number

	// Let's extract the available flows from the config
	// FromNull is handled separately, as it's considered an additional setting
	// to FromStrings and/or FromNumbers.
	var flows []string
	if !cfg.FromStrings.IsDisabled() {
		flows = append(flows, "string")
	}
	if !cfg.FromBools.IsDisabled() {
		flows = append(flows, "bool")
	}

	// If no conversion is allowed, return the clean boolean.
	if len(flows) == 0 {
		return v
	}

	// Randomly choose one conversion flow.
	switch flipping.FeelingLucky(flows, d.coin) {
	case "bool":
		// number from bool is possible only for 0, 1 values
		if v == 0 || v == 1 {
			return d.maybeNumberNilify(v, v != 0)
		}
		// if we only have `bool` flow, then for another number we just keep it clean
		if !slices.Contains(flows, "string") {
			return v
		}
		// otherwise: fallback to from-string logic
		fallthrough

	case "string":
		// TODO: support more configs
		_ = cfg.FromStrings.SpacingAllowed
		_ = cfg.FromStrings.ExponentNotationAllowed
		_ = cfg.FromStrings.CommasAllowed
		_ = cfg.FromStrings.RoundingAlgorithm

		// Convert the number to its string representation.
		// Optionally, if the config allowed spacing, commas, or exponent notation,
		// you could inject or remove them here.
		return d.maybeNumberNilify(v, fmt.Sprintf("%v", v))
	default:
		panic("unreachable")
	}
}

// maybeBoolNilify makes nil instead of bool respecting the Bool.FromNull config.
func (d *Dirtifier) maybeBoolNilify(v bool, actual any) any {
	if d.cfg.Bool.FromNull.IsDisabled() || d.keepItClean() {
		return actual
	}

	// let's make nilify logic even rare-er
	// As it's a very specific, rare, edge-case-ish logic.
	if d.coin.Flip() {
		return actual
	}

	if v == d.cfg.Bool.FromNull.Inverse {
		return nil
	}
	return actual
}

// maybeNumberNilify makes nil instead of number respecting the Number.FromNull config.
func (d *Dirtifier) maybeNumberNilify(v float64, actual any) any {
	if d.cfg.Bool.FromNull.IsDisabled() || d.keepItClean() {
		return actual
	}

	// let's make nilify logic even rare-er
	// As it's a very specific, rare, edge-case-ish logic.
	if d.coin.Flip() {
		return actual
	}

	if v == 0 {
		return nil
	}
	return actual
}
