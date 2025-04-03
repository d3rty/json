package cases_test

import (
	"strings"
	"testing"

	"github.com/d3rty/json/internal/cases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestIsCamelCase(t *testing.T) {
	assert.True(t, cases.Is("camelCase", cases.Camel), `"camelCase" should be camelCase`)
	assert.True(t, cases.Is("cAMEL", cases.Camel), `"cAMEL" should be camelCase`) // for now it's OK

	assert.False(t, cases.Is("", cases.Camel), `"" should not be camelCase`)
	assert.False(t, cases.Is("camelcase", cases.Camel), `"camelcase" should not be camelCase`)
	assert.False(t, cases.Is("CamelCase", cases.Camel), `"CamelCase" should not be camelCase`)
	assert.False(t, cases.Is("camel_case", cases.Camel), `"camel_case" should not be camelCase`)
	assert.False(t, cases.Is("camel-case", cases.Camel), `"camel-case" should not be camelCase`)
}

func TestIsPascalCase(t *testing.T) {
	// Valid PascalCase: starts with uppercase and becomes valid camelCase when first letter lowercased.
	assert.True(t, cases.Is("PascalCase", cases.Pascal), `"PascalCase" should be PascalCase`)

	assert.False(t, cases.Is("", cases.Pascal), `"" should not be PascalCase`)
	assert.False(t, cases.Is("PASCALCASE", cases.Pascal), `"PASCALCASE" should not be PascalCase`)
	assert.False(t, cases.Is("pascalcase", cases.Pascal), `"pascalcase" should not be PascalCase`)
	assert.False(t, cases.Is("pascalCase", cases.Pascal), `"pascalCase" should not be PascalCase`)
	assert.False(t, cases.Is("Pascal_Case", cases.Pascal), `"Pascal_Case" should not be PascalCase`)
	assert.False(t, cases.Is(
		"Pascalcase", cases.Pascal),
		`"Pascalcase" should not be PascalCase (missing internal uppercase)`,
	)
}

func TestIsSnakeCase(t *testing.T) {
	// Valid snake_case: contains underscore, all letters are lowercase.
	assert.True(t, cases.Is("snake_case", cases.Snake), `"snake_case" should be snake_case`)
	assert.True(t, cases.Is("_snake_case", cases.Snake), `"_snake_case" should be snake_case`)
	assert.True(t, cases.Is("_snake", cases.Snake), `"_snake" should be snake_case`)
	assert.True(t, cases.Is("_snake_", cases.Snake), `"_snake_" should be snake_case`)
	assert.True(t, cases.Is("snake_", cases.Snake), `"snake_" should be snake_case`)

	assert.False(t, cases.Is("", cases.Snake), `"" should not be snake_case`)
	assert.False(t, cases.Is("snakecase", cases.Snake), `"snakecase" should not be snake_case`)
	assert.False(t, cases.Is("Snake_case", cases.Snake), `"Snake_case" should not be snake_case`)
	assert.False(t, cases.Is("snakeCase", cases.Snake), `"snakeCase" should not be snake_case`)
	assert.False(t, cases.Is("snake-case", cases.Snake), `"snake-case" should not be snake_case`)
}

func TestIsTitleSnakeCase(t *testing.T) {
	// Valid cases.
	assert.True(t, cases.Is("Something_That_Ive_Never_Met", cases.TitleSnake), "Expected Title Snake Case")
	assert.True(t, cases.Is("Hello_World", cases.TitleSnake), "Expected Title Snake Case")
	assert.True(t, cases.Is("A_B_C", cases.TitleSnake), "Expected Title Snake Case with single letters")

	// Invalid cases.
	assert.False(t,
		cases.Is("NotTitleSnakeCase", cases.TitleSnake),
		"Missing underscore should return false",
	)
	assert.False(t,
		cases.Is("something_That_Ive_Never_met", cases.TitleSnake),
		"First segment starts with lowercase should return false",
	)
	assert.False(t,
		cases.Is("Something_THat_Ive_Never_met", cases.TitleSnake),
		"Segment with wrong casing should return false",
	)
	assert.False(t,
		cases.Is("Something__That_Ive_Never_met", cases.TitleSnake),
		"Empty segment due to consecutive underscores should return false",
	)
}

func TestIsKebabCase(t *testing.T) {
	// Valid kebab-case: contains hyphen, all letters lowercase.
	assert.True(t, cases.Is("kebab-case", cases.Kebab), `"kebab-case" should be kebab-case`)
	assert.True(t, cases.Is("-kebab-case", cases.Kebab), `"-kebab-case" should be kebab-case`)
	assert.True(t, cases.Is("-kebab", cases.Kebab), `"-kebab" should be kebab-case`)
	assert.True(t, cases.Is("kebab-", cases.Kebab), `"kebab-" should be kebab-case`)
	assert.True(t, cases.Is("-kebab-", cases.Kebab), `"-kebab-" should be kebab-case`)

	assert.False(t, cases.Is("", cases.Kebab), `"" should not be kebab-case`)
	assert.False(t, cases.Is("kebab", cases.Kebab), `"kebab" should not be kebab-case`)
	assert.False(t, cases.Is("Kebab-case", cases.Kebab), `"Kebab-case" should not be kebab-case`)
	assert.False(t, cases.Is("KebabCase", cases.Kebab), `"KebabCase" should not be kebab-case`)
	assert.False(t, cases.Is("kebabCase", cases.Kebab), `"kebabCase" should not be kebab-case`)
	assert.False(t, cases.Is("kebab_case", cases.Kebab), `"kebab_case" should not be kebab-case`)
}

func TestIsHeaderCase(t *testing.T) {
	// Valid header case: each word capitalized, separated by hyphens.
	assert.True(t, cases.Is("X-Header-Name", cases.Header), `"X-Header-Name" should be header case`)
	assert.True(t, cases.Is("Content-Type", cases.Header), `"Content-Type" should be header case`)

	assert.False(t, cases.Is("", cases.Header), `"" should not be header case`)
	assert.False(t, cases.Is("Host", cases.Header), `"Host" should not be header case`)
	assert.False(t, cases.Is("content-type", cases.Header), `"content-type" should not be header case`)
	assert.False(t, cases.Is("X-header-Name", cases.Header), `"X-header-Name" should not be header case`)
}

func TestIsComplexCase(t *testing.T) {
	// Strings that are in one of the complex cases should return true.
	assert.True(t, cases.IsComplexCase("camelCase"), "camelCase should be recognized as complex")
	assert.True(t, cases.IsComplexCase("PascalCase"), "PascalCase should be recognized as complex")
	assert.True(t, cases.IsComplexCase("snake_case"), "snake_case should be recognized as complex")
	assert.True(t, cases.IsComplexCase("kebab-case"), "kebab-case should be recognized as complex")
	assert.True(t, cases.IsComplexCase("X-Header-Name"), "X-Header-Name should be recognized as complex")
	assert.True(t, cases.IsComplexCase("Content-Type"), "Content-Type should be recognized as complex")
	assert.True(t,
		cases.IsComplexCase("Mixed-Case_with-mixed_separators"),
		"Mixed-Case_with-mixed_separators should not be recognized as complex",
	)

	// Strings that do not follow any of the naming conventions should return false.
	assert.False(t, cases.IsComplexCase(""), "Empty string should not be recognized as complex")
	assert.False(t, cases.IsComplexCase("Title"), "Title should not be recognized as complex")
	assert.False(t,
		cases.IsComplexCase("lowercase"),
		"lowercase should not be recognized as complex (missing internal uppercase for camel)",
	)

	assert.False(t, cases.IsComplexCase("UPPERCASE"), "UPPERCASE should not be recognized as complex")
}

func TestIsHybridCase(t *testing.T) {
	assert.True(t, cases.IsHybridCase("Mixed-Case_with-mixed_separators"), "should be hybrid case")
	assert.True(t, cases.IsHybridCase("mixed_Case"), "should be hybrid due to inconsistent casing")

	// Only hyphens or only underscores should not be considered hybrid.
	assert.False(t, cases.IsHybridCase("mixed-case"), "only hyphens should not be hybrid")
	assert.False(t, cases.IsHybridCase("mixed_case"), "only underscores should not be hybrid")
	assert.False(t, cases.IsHybridCase("X-Header-Name"), "header case is not hybrid")
	assert.False(t, cases.IsHybridCase(""), "empty string is not hybrid")
}

func TestTransformToHybridCase_NilRNG(t *testing.T) {
	input := "helloWorldTestFoo_BarBazOne_Two-three-Four-FiveSix"

	for range 100 {
		output := cases.TransformToHybridCase(input)

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

// SplitWordsSuite defines the suite for testing SplitWords.
type SplitWordsSuite struct {
	suite.Suite
}

// TestEmptyString tests that an empty input returns nil.
func (s *SplitWordsSuite) TestEmptyString() {
	result := cases.SplitWords("")
	s.Nil(result, "Expected nil for empty input")
}

// TestCamelCase tests splitting a camelCase string.
func (s *SplitWordsSuite) TestCamelCase() {
	result := cases.SplitWords("helloWorld")
	expected := []string{"hello", "World"}
	s.Equal(expected, result, "Should split camelCase words correctly")
}

// TestPascalCase tests splitting a PascalCase string.
func (s *SplitWordsSuite) TestPascalCase() {
	result := cases.SplitWords("HelloWorld")
	expected := []string{"Hello", "World"}
	s.Equal(expected, result, "Should split PascalCase words correctly")
}

// TestWithUnderscores tests splitting a string with underscores.
func (s *SplitWordsSuite) TestWithUnderscores() {
	result := cases.SplitWords("hello_world_test")
	expected := []string{"hello", "world", "test"}
	s.Equal(expected, result, "Should split underscore-delimited words correctly")
}

// TestWithHyphens tests splitting a string with hyphens.
func (s *SplitWordsSuite) TestWithHyphens() {
	result := cases.SplitWords("hello-world-test")
	expected := []string{"hello", "world", "test"}
	s.Equal(expected, result, "Should split hyphen-delimited words correctly")
}

// TestHybridCase tests a hybrid case with delimiters and camel case.
func (s *SplitWordsSuite) TestHybridCase() {
	result := cases.SplitWords("hello_World-TestExample")
	expected := []string{"hello", "World", "Test", "Example"}
	s.Equal(expected, result, "Should correctly handle hybrid cases")
}

// TestMultipleDelimiters tests handling of multiple consecutive delimiters.
func (s *SplitWordsSuite) TestMultipleDelimiters() {
	result := cases.SplitWords("hello__world--Test")
	expected := []string{"hello", "world", "Test"}
	s.Equal(expected, result, "Should ignore empty parts between consecutive delimiters")
}

// In order for 'go test' to run this suite, we need a Test function.
func TestSplitWordsSuite(t *testing.T) {
	suite.Run(t, new(SplitWordsSuite))
}
