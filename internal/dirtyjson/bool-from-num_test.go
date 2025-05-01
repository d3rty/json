package dirtyjson_test

import (
	"testing"

	"github.com/amberpixels/abu/maybe"
	"github.com/d3rty/json/internal/config"
	"github.com/d3rty/json/internal/dirtyjson"
	"github.com/stretchr/testify/assert"
)

// TestBoolFromNumberBinary tests the BoolFromNumberBinary algorithm.
func TestBoolFromNumberBinary(t *testing.T) {
	// Get the parser function for BoolFromNumberBinary
	parser := dirtyjson.GetBoolFromNumParser(config.BoolFromNumberBinary)

	// Test cases
	testCases := []struct {
		name     string
		input    float64
		expected maybe.Bool
	}{
		{"Zero returns false", 0, maybe.False()},
		{"One returns true", 1, maybe.True()},
		{"Negative number returns none", -1, maybe.NoneBool()},
		{"Positive number (not 1) returns none", 2, maybe.NoneBool()},
		{"Decimal number returns none", 0.5, maybe.NoneBool()},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parser(tc.input)
			assert.Equal(t, tc.expected, result, "Expected %v for input %v, got %v", tc.expected, tc.input, result)
		})
	}
}

// TestBoolFromNumberPositiveNegative tests the BoolFromNumberPositiveNegative algorithm.
func TestBoolFromNumberPositiveNegative(t *testing.T) {
	// Get the parser function for BoolFromNumberPositiveNegative
	parser := dirtyjson.GetBoolFromNumParser(config.BoolFromNumberPositiveNegative)

	// Test cases
	testCases := []struct {
		name     string
		input    float64
		expected maybe.Bool
	}{
		{"Zero returns false", 0, maybe.False()},
		{"Negative number returns false", -1, maybe.False()},
		{"Large negative number returns false", -1000, maybe.False()},
		{"Small positive number returns true", 0.1, maybe.True()},
		{"One returns true", 1, maybe.True()},
		{"Large positive number returns true", 1000, maybe.True()},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parser(tc.input)
			assert.Equal(t, tc.expected, result, "Expected %v for input %v, got %v", tc.expected, tc.input, result)
		})
	}
}

// TestBoolFromNumberSignOfOne tests the BoolFromNumberSignOfOne algorithm.
func TestBoolFromNumberSignOfOne(t *testing.T) {
	// Get the parser function for BoolFromNumberSignOfOne
	parser := dirtyjson.GetBoolFromNumParser(config.BoolFromNumberSignOfOne)

	// Test cases
	testCases := []struct {
		name     string
		input    float64
		expected maybe.Bool
	}{
		{"Negative one returns false", -1, maybe.False()},
		{"One returns true", 1, maybe.True()},
		{"Zero returns none", 0, maybe.NoneBool()},
		{"Negative number (not -1) returns none", -2, maybe.NoneBool()},
		{"Positive number (not 1) returns none", 2, maybe.NoneBool()},
		{"Decimal number returns none", 0.5, maybe.NoneBool()},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parser(tc.input)
			assert.Equal(t, tc.expected, result, "Expected %v for input %v, got %v", tc.expected, tc.input, result)
		})
	}
}

// TestBoolFromNumberUndefined tests the BoolFromNumberUndefined algorithm.
func TestBoolFromNumberUndefined(t *testing.T) {
	// Get the parser function for BoolFromNumberUndefined
	parser := dirtyjson.GetBoolFromNumParser(config.BoolFromNumberUndefined)

	// The parser should be nil
	assert.Nil(t, parser, "Parser for BoolFromNumberUndefined should be nil")
}

// TestAllBoolFromNumberAlgs tests all BoolFromNumberAlg types.
func TestAllBoolFromNumberAlgs(t *testing.T) {
	// Get all available BoolFromNumberAlg types
	algs := config.ListAvailableBoolFromNumberAlgs()

	// Make sure we have all the expected algorithms
	expectedAlgs := []config.BoolFromNumberAlg{
		config.BoolFromNumberBinary,
		config.BoolFromNumberPositiveNegative,
		config.BoolFromNumberSignOfOne,
	}

	// Check that all expected algorithms are in the list
	for _, expected := range expectedAlgs {
		found := false
		for _, alg := range algs {
			if alg == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected algorithm %v not found in available algorithms", expected)
	}

	// Check that we have parsers for all available algorithms
	for _, alg := range algs {
		parser := dirtyjson.GetBoolFromNumParser(alg)
		assert.NotNil(t, parser, "Parser for algorithm %v should not be nil", alg)
	}
}
