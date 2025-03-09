package dirtytesting

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"unicode"

	"github.com/d3rty/json/internal/cases"
	"github.com/d3rty/json/internal/config"
)

// Dirtyfier makes dirty JSONs from clean ones.
type Dirtyfier struct {
	// threshold is a number from 0.0 to 1.0 that sets how much "dirty" the result will end up.
	// E.g. 1.0 means that 100% of possible fields will be dirtified.
	//      0.0 means that 0% of possible fields will be dirtified (so result will remain clean).
	threshold float64

	// rng is the random generator that returns value [0-1) that is compared to the threshold.
	// it allows to make random decision for each field to be diritifed
	rng *rand.Rand

	// cfg stands for dirty Config: how specifically we do dirty.
	cfg *config.Config
}

// NewDirtyfiers creates a new dirtyfier
func NewDirtyfier(threshold float64, cfg *config.Config, rngArg ...*rand.Rand) *Dirtyfier {
	var rng *rand.Rand
	if len(rngArg) > 0 {
		rng = rngArg[0]
	} else {
		rng = newRng()
	}

	return &Dirtyfier{
		rng:       rng,
		cfg:       cfg,
		threshold: threshold,
	}
}

// keepItClean returns true/false depending on current rand generation and the threshold
// keepItClean returns true if we should omit dirtyfing, and false if we should do dirtyfing.
func (d *Dirtyfier) keepItClean() bool {
	return d.rng.Float64() >= d.threshold
}

// flipTheCount returns true or false with 50% chance
func (d *Dirtyfier) flipTheCoin() bool {
	return d.rng.Float64() >= 0.5
}

func (d *Dirtyfier) randomCase(s string) string {
	// Optionally, seed the random number generator once in your main function or init block.
	// rand.Seed(time.Now().UnixNano())

	// Convert the string to a slice of runes to properly handle Unicode characters.
	runes := []rune(s)
	for i, r := range runes {
		// With 50% chance convert to lower case, else upper case.
		if d.flipTheCoin() {
			runes[i] = unicode.ToLower(r)
		} else {
			runes[i] = unicode.ToUpper(r)
		}
	}
	return string(runes)
}

// Dirtify mixes values applying random transformations.
func (d *Dirtyfier) Dirtify(val any) any {
	switch v := val.(type) {
	case map[string]any:
		// Process map values.
		newMap := make(map[string]any)
		for key, elem := range v {
			newMap[d.makeDirtyKey(key)] = d.Dirtify(elem)
		}
		return newMap
	case []any:
		// Process each element of the slice.
		for i, elem := range v {
			v[i] = d.Dirtify(elem)
		}
		return v
	case bool:
		// With probability, transform booleans to allowed string representations.
		return d.makeDirtyBool(v)
	case float64:
		// With probability, transform numbers to string.
		return d.makeDirtyNumber(v)
	case string:
		// strings are not dirtyfied yet
		return v
	default:
		// For other types (strings, nil, etc.), leave unchanged.
		return v
	}
}

// makeDirtyKey makes a dirty key from given clean key
func (d *Dirtyfier) makeDirtyKey(key string) string {
	if !d.cfg.FlexKeys.Allowed || !d.cfg.FlexKeys.CaseInsensitive && !d.cfg.FlexKeys.ChameleonCase {
		return key
	}
	if d.keepItClean() {
		return key
	}

	if cases.IsComplexCase(key) && d.cfg.FlexKeys.ChameleonCase && d.flipTheCoin() {
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

		convertTo := allCases[rand.Intn(len(allCases))]
		if convertTo == cases.Hybrid {
			return cases.TransformToHybridCase(key, d.threshold)
		}

		return cases.TransformTo(key, convertTo)
	}

	if d.cfg.FlexKeys.CaseInsensitive && d.flipTheCoin() {
		// Let's mix the case (make it upper/lower/title/etc)

		// We have to shuffle and try them until we get first transformation that makes sense
		// This is made in case key is already Upper so ToUpper transform will be ignored.
		transformations := []func(string) string{
			strings.ToUpper,
			strings.ToLower,
			strings.ToTitle,

			// todo: cases.ToRandomTitle
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

// makeDirtyBool makes a dirty bool from given clean bool
func (d *Dirtyfier) makeDirtyBool(v bool) (result any) {
	if !d.cfg.Bool.Allowed || d.keepItClean() {
		return v
	}

	cfg := d.cfg.Bool

	// Let's extract the available flows from the config
	// FromNull is handled separately, as it's considered additional setting
	// to FromStrings and/or FromNumbers.
	var flows []string
	if cfg.FromStrings.Allowed {
		flows = append(flows, "string")
	}
	if cfg.FromNumbers.Allowed {
		flows = append(flows, "number")
	}
	// If no conversion is allowed, return the clean boolean.
	if len(flows) == 0 {
		return v
	}

	// TODO: support .FallbackValue as it's not so simple

	// Randomly choose one conversion flow.
	var numberToBeStringified bool
	switch feelingLucky(d.rng, flows) {
	case "string":
		cfg := d.cfg.Bool.FromStrings

		// if we respect numbers logic then with 50% chance generate a stringish number
		// e.g. "1", "0", etc
		// Here we can fallthrough into "number" case if we simmply flip the true coin
		// or we have no option (custom strings are disabled in config)
		if cfg.RespectFromNumbersLogic {
			customStringsDisabled := (!v && len(cfg.CustomListForFalse) == 0 ||
				v && len(cfg.CustomListForTrue) == 0)

			if d.flipTheCoin() || customStringsDisabled {
				numberToBeStringified = true
			}
		}

		if !numberToBeStringified {
			if v {
				sTrue := feelingLucky(d.rng, cfg.CustomListForTrue)
				if cfg.CaseInsensitive && d.flipTheCoin() {
					sTrue = d.randomCase(sTrue)
				}
				return d.maybeBoolNilify(v, sTrue)
			}

			sFalse := feelingLucky(d.rng, cfg.CustomListForFalse)
			if cfg.CaseInsensitive && d.flipTheCoin() {
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
			if v {
				intBoolResult = d.rng.Intn(1000) + 1
			} else {
				intBoolResult = -d.rng.Intn(1000)
			}
		case config.BoolFromNumberSignOfOne:
			if v {
				intBoolResult = 1
			} else {
				intBoolResult = -1
			}
		}

		if numberToBeStringified {
			return d.maybeBoolNilify(v, strconv.Itoa(intBoolResult))
		}
		return d.maybeBoolNilify(v, intBoolResult)
	default:
		panic("unreachable")
	}
}

// makeDirtyNumber makes a dirty number from given clean number
func (d *Dirtyfier) makeDirtyNumber(v float64) any {
	if !d.cfg.Number.Allowed || d.keepItClean() {
		return v
	}

	cfg := d.cfg.Number

	// Let's extract the available flows from the config
	// FromNull is handled separately, as it's considered additional setting
	// to FromStrings and/or FromNumbers.
	var flows []string
	if cfg.FromStrings.Allowed {
		flows = append(flows, "string")
	}
	if cfg.FromBools.Allowed {
		flows = append(flows, "bool")
	}

	// If no conversion is allowed, return the clean boolean.
	if len(flows) == 0 {
		return v
	}

	// Randomly choose one conversion flow.
	switch feelingLucky(d.rng, flows) {
	case "string":
		// TODO: support more configs
		_ = cfg.FromStrings.SpacingAllowed
		_ = cfg.FromStrings.ExponentNotationAllowed
		_ = cfg.FromStrings.CommasAllowed
		_ = cfg.FromStrings.FloatishAllowed

		// Convert number to its string representation.
		// Optionally, if the config allowed spacing, commas, or exponent notation,
		// you could inject or remove them here.
		return d.maybeNumberNilify(v, fmt.Sprintf("%v", v))
	case "bool":

		return d.maybeNumberNilify(v, v != 0)
	default:
		panic("unreachable")
	}
}

// maybeBoolNilify makes a nil instead of bool respecting the Bool.FromNull config
func (d *Dirtyfier) maybeBoolNilify(v bool, actual any) any {
	if !d.cfg.Bool.FromNull.Allowed || d.keepItClean() {
		return actual
	}

	// let's make nilify logic even rare-er
	// As it's a very specific, rare, edge-casy logic.
	if d.flipTheCoin() {
		return actual
	}

	if v == d.cfg.Bool.FromNull.Inverse {
		return nil
	}
	return actual
}

// maybeNumberNilify makes a nil instead of number respecting the Number.FromNull config
func (d *Dirtyfier) maybeNumberNilify(v float64, actual any) any {
	if !d.cfg.Bool.FromNull.Allowed || d.keepItClean() {
		return actual
	}

	// let's make nilify logic even rare-er
	// As it's a very specific, rare, edge-casy logic.
	if d.flipTheCoin() {
		return actual
	}

	if v == 0 {
		return nil
	}
	return actual
}
