package dirtytesting

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/d3rty/json/internal/cases"
	"github.com/d3rty/json/internal/config"
)

type Mixer struct {
	threshold float64
	rng       *rand.Rand
	cfg       *config.Config
}

func NewMixer(threshold float64, rngArg ...*rand.Rand) *Mixer {
	var rng *rand.Rand
	if len(rngArg) > 0 {
		rng = rngArg[0]
	} else {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	return &Mixer{
		rng:       rng,
		cfg:       RandomConfig(rng),
		threshold: threshold,
	}
}

// Mix mixes values applying random transformations.
func (m *Mixer) Mix(val any) any {
	switch v := val.(type) {
	case map[string]any:
		// Process map values.
		newMap := make(map[string]any)
		for key, elem := range v {
			newMap[m.MixKey(key)] = m.Mix(elem)
		}
		return newMap
	case []any:
		// Process each element of the slice.
		for i, elem := range v {
			v[i] = m.Mix(elem)
		}
		return v
	case bool:
		// With probability, transform booleans to allowed string representations.
		return m.MixBool(v)
	case float64:
		// With probability, transform numbers to string.
		return m.MixNumber(v)
	default:
		// For other types (strings, nil, etc.), leave unchanged.
		return v
	}
}

func (m *Mixer) MixKey(key string) string {
	if !m.cfg.FlexKeys.Allowed || !m.cfg.FlexKeys.CaseInsensitive && !m.cfg.FlexKeys.ChameleonCase {
		return key
	}

	if cases.IsComplexCase(key) && m.cfg.FlexKeys.ChameleonCase && m.rng.Float64() < m.threshold {
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
			return cases.TransformToHybridCase(key, m.threshold)
		}

		return cases.TransformTo(key, convertTo)
	}

	if m.cfg.FlexKeys.CaseInsensitive && m.rng.Float64() < m.threshold {
		// Let's mix the case (make it upper/lower/title/etc)

		transformations := []func(string) string{
			strings.ToUpper,
			strings.ToLower,
			strings.ToTitle,

			// todo: cases.ToRandomTitle
		}

		// We have to shuffle and try them until we get first transformation that makes sense
		// This is made in case key is already Upper so ToUpper transform will be ignored.

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

func (m *Mixer) MixBool(v bool) any {
	if !m.cfg.Bool.Allowed {
		return v
	}

	// We won to choose the clean bool
	if m.rng.Float64() >= m.threshold {
		return v
	}

	var flows []string
	if m.cfg.Bool.FromStrings.Allowed {
		flows = append(flows, "string")
	}
	if m.cfg.Bool.FromNumbers.Allowed {
		flows = append(flows, "number")
	}
	if m.cfg.Bool.FromNull.Allowed {
		flows = append(flows, "null")
	}
	// If no conversion is allowed, return the clean boolean.
	if len(flows) == 0 {
		return v
	}

	// Randomly choose one conversion flow.
	chosenFlow := flows[m.rng.Intn(len(flows))]
	switch chosenFlow {
	case "string":
		// For string conversion, choose a representation based on val.
		var list []string
		if v {
			if len(m.cfg.Bool.FromStrings.CustomListForTrue) == 0 {
				return v
			}
			// Make a copy of the list.
			list = append([]string(nil), m.cfg.Bool.FromStrings.CustomListForTrue...)
		} else {
			if len(m.cfg.Bool.FromStrings.CustomListForFalse) == 0 {
				return v
			}
			list = append([]string(nil), m.cfg.Bool.FromStrings.CustomListForFalse...)
		}
		// Shuffle the list to randomize the selection.
		m.rng.Shuffle(len(list), func(i, j int) {
			list[i], list[j] = list[j], list[i]
		})
		return list[0]
	case "number":
		// For number conversion, simulate a bit of variety.
		// For example, if true return either 1 or 500; if false, return 0 or -100.
		if v {
			if m.rng.Intn(2) == 0 {
				return 1
			}
			return 500
		} else {
			if m.rng.Intn(2) == 0 {
				return 0
			}
			return -100
		}
	case "null":
		// TODO: should be only if config allows it. so if null is false, etc
		// For null conversion, always return nil.
		return nil

	default:
		panic("unreachable")
	}
}

func (m *Mixer) MixNumber(v float64) any {
	if !m.cfg.Number.Allowed {
		return v
	}

	// We won to choose the clean bool
	if m.rng.Float64() >= m.threshold {
		return v
	}

	var flows []string
	if m.cfg.Number.FromStrings.Allowed {
		flows = append(flows, "string")
	}
	if m.cfg.Number.FromBools.Allowed {
		flows = append(flows, "bool")
	}
	if m.cfg.Number.FromNull.Allowed {
		flows = append(flows, "null")
	}
	// If no conversion is allowed, return the clean boolean.
	if len(flows) == 0 {
		return v
	}

	// Randomly choose one conversion flow.
	chosenFlow := flows[m.rng.Intn(len(flows))]
	switch chosenFlow {
	case "string":
		// Convert number to its string representation.
		// Optionally, if the config allowed spacing, commas, or exponent notation,
		// you could inject or remove them here.
		s := fmt.Sprintf("%v", v)
		return s
	case "bool":
		// Convert number to a boolean.
		// Typically, non-zero becomes true and zero becomes false.
		if v == 0 {
			return false
		}
		return true
	case "null":
		// Simulate a null value.
		return nil
	default:
		return v
	}
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

	// Make a copy of the master list so we can shuffle it.
	shuffled := make([]string, len(dict))
	copy(shuffled, dict)
	// Shuffle the master copy.
	r.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Return the first count elements.
	return shuffled[:count]
}
