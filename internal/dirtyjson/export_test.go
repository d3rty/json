package dirtyjson

import (
	"github.com/amberpixels/k1/maybe"
	"github.com/d3rty/json/internal/config"
)

// GetBoolFromNumParser returns the parser function for the given BoolFromNumberAlg.
// This function is exported only for testing purposes.
func GetBoolFromNumParser(alg config.BoolFromNumberAlg) func(float64) maybe.Bool {
	return parsersBoolFromNum[alg]
}

// LimitedStr exports limitedStr for testing purposes.
func LimitedStr(s string, limitArg ...int) string {
	return limitedStr(s, limitArg...)
}

// GetStringBetweenQuotes exports getStringBetweenQuotes for testing purposes.
func GetStringBetweenQuotes(data []byte) (string, error) {
	return getStringBetweenQuotes(data)
}

// NormalizeJSONKey exports normalizeJSONKey for testing purposes.
func NormalizeJSONKey(key string) string {
	return normalizeJSONKey(key)
}
