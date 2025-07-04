package cases

import (
	"slices"
	"strings"
	"unicode"
)

// Case defines the target naming convention.
type Case string

const (
	Snake      Case = "snake"       // e.g. "hello_world"
	Camel      Case = "camel"       // e.g. "helloWorld"
	Pascal     Case = "pascal"      // e.g. "HelloWorld"
	Kebab      Case = "kebab"       // e.g. "hello-world"
	Header     Case = "header"      // e.g. "Hello-World"
	TitleSnake Case = "title_snake" // e.g. "Hello_World"
	Hybrid     Case = "hybrid"      // e.g. "Hello_beautiful-WorldHere"
)

// hasValuableSymbols returns true if the given string has at least one valuable symbol.
// Valuable symbols are any characters except delimiters (hyphen, underscore, space).
func hasValuableSymbols(s string) bool {
	if s == "" {
		return false
	}

	return strings.ContainsFunc(s, func(r rune) bool {
		return r != '-' && r != '_' && r != ' '
	})
}

// isSnakeCase returns true if s is snake_case.
// It requires that s contains at least one underscore and that
// all alphabetic characters are lowercase.
func isSnakeCase(s string) bool {
	if !hasValuableSymbols(s) { // no valuable symbols are always considered as `false`
		return false
	}
	if !strings.Contains(s, "_") {
		return false
	}
	for _, r := range s {
		if r == '_' {
			continue
		}
		if unicode.IsLetter(r) && !unicode.IsLower(r) {
			return false
		}
		// snake_case means lower letters, digits and underscores only
	}
	return true
}

// isCamelCase returns true if s is camelCase.
// It checks that s starts with a lowercase letter, contains no underscores or hyphens,
// and has at least one uppercase letter (after the first character).
func isCamelCase(s string) bool {
	if !hasValuableSymbols(s) { // no valuable symbols are always considered as `false`
		return false
	}
	runes := []rune(s)

	// Must start with a lowercase letter.
	if !unicode.IsLower(runes[0]) {
		return false
	}
	// Should not contain underscores or hyphens.
	if strings.Contains(s, "_") || strings.Contains(s, "-") {
		return false
	}

	// Must contain at least one uppercase letter beyond the first character.
	return slices.ContainsFunc(runes[1:], unicode.IsUpper)
}

// isPascalCase returns true if s is PascalCase.
// It ensures that s starts with an uppercase letter and that,
// after lowercasing the first letter, the result is a valid camelCase.
func isPascalCase(s string) bool {
	if !hasValuableSymbols(s) { // no valuable symbols are always considered as `false`
		return false
	}
	runes := []rune(s)
	if !unicode.IsUpper(runes[0]) {
		return false
	}

	// all uppers are not pascal case
	nUpper := 0
	for _, r := range runes {
		if !unicode.IsUpper(r) {
			break
		}
		nUpper++
	}
	if len(runes) == nUpper {
		return false
	}

	// Convert the first letter to lowercase and check if it's camelCase.
	lowered := string(unicode.ToLower(runes[0])) + string(runes[1:])
	return isCamelCase(lowered)
}

// isHeaderCase returns true if s is in the header case.
// Header case means the string is split by hyphens, and for each segment,
// the first letter is uppercase (if it is a letter), and all further letters are lowercase.
// HeaderCase considers we have at least one hyphen (so two words or more words).
func isHeaderCase(s string) bool {
	if s == "" {
		return false
	}

	// Split the string into parts using hyphen as a separator.
	parts := strings.Split(s, "-")
	if len(parts) <= 1 {
		return false
	}

	for _, part := range parts {
		// each part must have at least one valuable symbol
		if !hasValuableSymbols(part) {
			return false
		}
		runes := []rune(part)
		// The first rune must be uppercase if it's a letter.
		if unicode.IsLetter(runes[0]) && !unicode.IsUpper(runes[0]) {
			return false
		}
		// Check the rest of the runes: if they are letters, they should be lowercase.
		for _, r := range runes[1:] {
			if unicode.IsLetter(r) && !unicode.IsLower(r) {
				return false
			}
		}
	}
	return true
}

// isKebabCase returns true if s is a kebab-case.
// It requires that s contains at least one hyphen and that
// all alphabetic characters are lowercase.
func isKebabCase(s string) bool {
	if !hasValuableSymbols(s) { // no valuable symbols are always considered as `false`
		return false
	}
	if !strings.Contains(s, "-") {
		return false
	}
	for _, r := range s {
		if r == '-' {
			continue
		}
		if unicode.IsLetter(r) && !unicode.IsLower(r) {
			return false
		}
	}
	return true
}

// isTitleSnakeCase returns true if s is in Title Snake Case.
// Title Snake Case means the string contains underscores and each segment (separated by underscores)
// starts with an uppercase letter and is followed by only lowercase letters.
func isTitleSnakeCase(s string) bool {
	if !hasValuableSymbols(s) { // no valuable symbols are always considered as `false`
		return false
	}

	parts := strings.Split(s, "_")
	if len(parts) == 1 { // Must contain at least one underscore.
		return false
	}
	// All parts must be non-empty and follow the title rule.
	for _, part := range parts {
		if part == "" {
			return false
		}
		runes := []rune(part)
		// The first rune must be uppercase.
		if !unicode.IsUpper(runes[0]) {
			return false
		}
		// The remaining runes should be lowercase letters (if they are letters).
		for _, r := range runes[1:] {
			if unicode.IsLetter(r) && !unicode.IsLower(r) {
				return false
			}
		}
	}
	return true
}

// Is returns true if s is in the target case.
func Is(s string, target Case) bool {
	switch target {
	case TitleSnake:
		return isTitleSnakeCase(s)
	case Snake:
		return isSnakeCase(s)
	case Camel:
		return isCamelCase(s)
	case Pascal:
		return isPascalCase(s)
	case Kebab:
		return isKebabCase(s)
	case Header:
		return isHeaderCase(s)
	case Hybrid:
		return IsHybridCase(s)
	default:
		return false
	}
}

// IsComplexCase returns true if s is in one of the recognized naming conventions:
// PascalCase, camelCase, snake_case, kebab-case, Header-Case, or hybrid_Case.
func IsComplexCase(s string) bool {
	if !hasValuableSymbols(s) { // no valuable symbols are always considered as `false`
		return false
	}

	if strings.Contains(s, "-") || strings.Contains(s, "_") {
		return true
	}

	// because of `-` and `_` check, only CamelCase and PascalCase remained not covered
	return isCamelCase(s) || isPascalCase(s)
}

// IsHybridCase returns true if the string contains a mix of separators (e.g., both "-" and "_")
// and a mix of casing that does not conform strictly to one of the standard conventions.
func IsHybridCase(s string) bool {
	if !hasValuableSymbols(s) { // no valuable symbols are always considered as `false`
		return false
	}

	// Check if it contains at least two different types of separators.
	hasHyphen := strings.Contains(s, "-")
	hasUnderscore := strings.Contains(s, "_")
	if !hasHyphen && !hasUnderscore {
		return false
	}
	// If both hyphen and underscore are present, it's a hybrid.
	if hasHyphen && hasUnderscore {
		return true
	}

	// If only one separator is present, check if the parts are inconsistently cased.
	var sep string
	if hasHyphen {
		sep = "-"
	} else {
		sep = "_"
	}

	parts := strings.Split(s, sep)
	// If the parts don't have a consistent case pattern, we can consider it hybrid.
	// For simplicity, if at least one part starts with uppercase and at least one with lowercase, flag it.
	var hasUpper, hasLower bool
	for _, part := range parts {
		if part == "" {
			continue
		}
		r := []rune(part)[0]
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
	}
	return hasUpper && hasLower
}

// TransformTo transforms a given string into a target Case.
// It supports determined cases only, but not Hybrid. For Hybrid use TransformToHybrid.
func TransformTo(s string, target Case) string {
	words := SplitWords(s)
	switch target {
	case Snake:
		return strings.Join(lowerWords(words), "_")
	case Camel:
		if len(words) == 0 {
			return ""
		}
		return strings.ToLower(words[0]) + joinCapitalized(words[1:])
	case Pascal:
		return joinCapitalized(words)
	case Kebab:
		return strings.Join(lowerWords(words), "-")
	case Header:
		return strings.Join(capitalizeWords(words), "-")
	case TitleSnake:
		return strings.Join(capitalizeWords(words), "_")
	case Hybrid:
		panic("TransformTo can only accept determined cases. For CaseHybrid use TransformToHybrid")
	default:
		// If unknown, return the input unmodified.
		return s
	}
}

// separatorRunes are a list of runes used for separation in hybrid case
// '\x00' represents the empty rune. E.g., It's used for `camelCase` separation.
const separatorRunes = "-_ \x00"

// SplitWords attempts to split an input string into words.
// It handles strings that use underscores, hyphens, or camel/pascal style,
// and works well for hybrid cases (mixing these conventions).
func SplitWords(s string) []string {
	if !hasValuableSymbols(s) { // nothing to split
		return []string{}
	}

	// If the string contains underscores or hyphens,
	// split by these delimiters.
	if strings.ContainsAny(s, separatorRunes) {
		// Split on both '_' and '-'
		parts := strings.FieldsFunc(s, func(r rune) bool {
			return r == '_' || r == '-' || r == ' ' || r == '\x00'
		})
		var words []string
		// For each part, check if it has a mixed case (camel/Pascal style).
		// If yes, further split it; otherwise, use the part as-is.
		for _, part := range parts {
			if part == "" {
				continue
			}
			if hasMixedCase(part) {
				words = append(words, splitCamelCase(part)...)
			} else {
				words = append(words, part)
			}
		}
		return words
	}

	// Otherwise, assume the string is in camelCase or PascalCase.
	return splitCamelCase(s)
}

// hasMixedCase returns true if s contains at least one uppercase and one lowercase letter.
func hasMixedCase(s string) bool {
	var hasUpper, hasLower bool
	for _, r := range s {
		if unicode.IsUpper(r) {
			hasUpper = true
		} else if unicode.IsLower(r) {
			hasLower = true
		}
		if hasUpper && hasLower {
			return true
		}
	}
	return false
}

// splitCamelCase splits a camelCase or PascalCase string into its words.
func splitCamelCase(s string) []string {
	var words []string
	var lastIdx int
	runes := []rune(s)
	for i := 1; i < len(runes); i++ {
		// If the current rune is uppercase and the previous rune is lowercase or digit,
		// consider it as a boundary.
		if unicode.IsUpper(runes[i]) && (unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1])) {
			words = append(words, string(runes[lastIdx:i]))
			lastIdx = i
		}
	}
	words = append(words, string(runes[lastIdx:]))
	return words
}

// lowerWords returns a new slice with all words in lowercase.
func lowerWords(words []string) []string {
	out := make([]string, len(words))
	for i, w := range words {
		out[i] = strings.ToLower(w)
	}
	return out
}

// capitalizeWords returns a new slice with the first letter capitalized and the rest lowercased.
func capitalizeWords(words []string) []string {
	out := make([]string, len(words))
	for i, w := range words {
		if w == "" {
			out[i] = ""
		} else {
			out[i] = strings.ToUpper(string([]rune(w)[0])) + strings.ToLower(w[1:])
		}
	}
	return out
}

// joinCapitalized concatenates words by capitalizing each word.
func joinCapitalized(words []string) string {
	return strings.Join(capitalizeWords(words), "")
}
