package dirtyjson

import (
	"errors"
	"strings"
)

const (
	maxMessageLength = 50
)

func limitedStr(s string, limitArg ...int) string {
	var limit = maxMessageLength
	if len(limitArg) > 0 {
		limit = limitArg[0]
	}

	if len(s) > limit {
		return s[0:limit] + "…"
	}

	return s
}

// getStringBetweenQuotes returns the string between the quotes.
// It returns an error if the string is not valid.
// E.g. `"hello"` returns `hello`.
// `"something`, `something`, `something "else"`, etc  will fail.
func getStringBetweenQuotes(data []byte) (string, error) {
	if data[0] != '"' {
		return "", errors.New("quoted string must start with a quote")
	}
	if len(data) < 2 || data[len(data)-1] != '"' {
		return "", errors.New("invalid string value")
	}

	s := string(data[1 : len(data)-1])
	s = strings.TrimSpace(s)

	return s, nil
}

// normalizeJSONKeys normalizes given string:
// makes it lowercase + removes _,-, spaces.
func normalizeJSONKey(key string) string {
	if key == "" {
		return key
	}

	var sb strings.Builder
	sb.Grow(len(key)) // preallocate memory
	for i := range len(key) {
		c := key[i]
		// Skip underscores, dashes, and spaces.
		if c == '_' || c == '-' || c == ' ' {
			continue
		}
		// Convert uppercase letters to lowercase.
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		sb.WriteByte(c)
	}

	return sb.String()
}
