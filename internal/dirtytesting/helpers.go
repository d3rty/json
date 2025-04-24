package dirtytesting

import (
	"encoding/json"
	"unicode"

	"github.com/d3rty/json/internal/cases"
	"github.com/d3rty/json/internal/flipping"
)

// StructToMap converts any struct to a map[string]any via JSON round-trip.
func StructToMap(s any) map[string]any {
	var m map[string]any
	b, err := json.Marshal(s)
	if err != nil {
		panic("structToMap: failed to marshal struct " + err.Error())
	}
	if err := json.Unmarshal(b, &m); err != nil {
		panic("structToMap: failed to unmarshal struct " + err.Error())
	}
	return m
}

// separatorRunes are a list of runes used for separation in hybrid case
// '\x00' represents the empty rune. E.g., It's used for `camelCase` separation.
const separatorRunes = "-_ \x00"

// TransformToHybridCase transforms the input string s into a hybrid case string.
// It uses randomness to decide, for each gap between words, whether to insert an underscore ("_"),
// a hyphen ("-"), or no separator at all.
// When no separator is chosen, if the last character of the previous word and the first character of the next word
// are both lowercase (which might merge the words indistinguishably),
// then the empty separator is overridden with either a hyphen or underscore.
// The forced choice uses hyphenRatio: with probability hyphenRatio a hyphen is used, otherwise an underscore.
// The function accepts an optional RNG argument (variadic); if none is provided, a default RNG is used.
func TransformToHybridCase(s string, coinArg ...*flipping.Coin) string {
	words := cases.SplitWords(s)
	if len(words) == 0 {
		return s
	}

	coin := flipping.MaybeNewCoin(coinArg...)

	// Start with the first word as-is.
	result := words[0]
	for i := 1; i < len(words); i++ {
		sep := flipping.FeelingLucky([]rune(separatorRunes), coin)

		// If no separator was chosen, check if joining the words would merge them indistinguishably.
		if sep == '\x00' {
			prevRunes := []rune(result)
			nextRunes := []rune(words[i])
			if len(prevRunes) > 0 && len(nextRunes) > 0 {
				lastRune := prevRunes[len(prevRunes)-1]
				firstRune := nextRunes[0]
				// If both are lowercase, force a separator.
				if unicode.IsLower(lastRune) && unicode.IsLower(firstRune) {
					sep = rune(separatorRunes[coin.Rng().Intn(2)])
				}
			}
		}

		result += string(sep) + words[i]
	}
	return result
}
