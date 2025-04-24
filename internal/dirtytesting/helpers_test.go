package dirtytesting_test

import (
	"strings"
	"testing"

	"github.com/d3rty/json/internal/cases"
	"github.com/d3rty/json/internal/dirtytesting"
	"github.com/stretchr/testify/assert"
)

func TestTransformToHybridCase_NilRNG(t *testing.T) {
	input := "helloWorldTestFoo_BarBazOne_Two-three-Four-FiveSix"

	for range 100 {
		output := dirtytesting.TransformToHybridCase(input)

		// We cannot predict the exact output here, but we can verify some properties:
		// - The output should be non-empty.
		// - It should contain at least one separator, either "-" or "_". (might be flaky. fix later if needed)
		assert.NotEmpty(t, output, "Output should not be empty")
		assert.True(t, strings.Contains(output, "-") || strings.Contains(output, "_"),
			"Output should contain at least one hyphen or underscore as separator: "+output)

		// Additionally, we can check that the words are transformed in either lower or title case.
		words := cases.SplitWords(output)
		assert.Len(t, words, 12, words)
		for _, word := range words {
			// Each word should have its first letter either uppercase or lowercase.
			if len(word) > 0 {
				first := word[0]
				assert.True(t, (first >= 'A' && first <= 'Z') || (first >= 'a' && first <= 'z'),
					"Each word should start with an alphabetic character")
			}
		}
	}
}
