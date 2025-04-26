package dirtyjson_test

import (
	"testing"

	"github.com/d3rty/json/internal/dirtyjson"
)

func TestLimitedStr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		limit    []int
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "string shorter than default limit",
			input:    "short string",
			expected: "short string",
		},
		{
			name:     "string equal to default limit",
			input:    "this string is exactly 50 chars long 1234567890123",
			expected: "this string is exactly 50 chars long 1234567890123",
		},
		{
			name:     "string longer than default limit",
			input:    "this string is longer than 50 characters and should be truncated",
			expected: "this string is longer than 50 characters and shoul" + "…",
		},
		{
			name:     "custom limit - shorter string",
			input:    "short",
			limit:    []int{10},
			expected: "short",
		},
		{
			name:     "custom limit - longer string",
			input:    "this string is longer than 10 characters",
			limit:    []int{10},
			expected: "this strin" + "…",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dirtyjson.LimitedStr(tt.input, tt.limit...)
			if result != tt.expected {
				t.Errorf("LimitedStr(%q, %v) = %q, want %q", tt.input, tt.limit, result, tt.expected)
			}
		})
	}
}

func TestGetStringBetweenQuotes(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expected    string
		expectError bool
	}{
		{
			name:        "valid quoted string",
			input:       []byte(`"hello"`),
			expected:    "hello",
			expectError: false,
		},
		{
			name:        "valid quoted string with spaces",
			input:       []byte(`"  hello world  "`),
			expected:    "hello world",
			expectError: false,
		},
		{
			name:        "missing opening quote",
			input:       []byte(`hello"`),
			expectError: true,
		},
		{
			name:        "missing closing quote",
			input:       []byte(`"hello`),
			expectError: true,
		},
		{
			name:        "empty quoted string",
			input:       []byte(`""`),
			expected:    "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dirtyjson.GetStringBetweenQuotes(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("GetStringBetweenQuotes(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("GetStringBetweenQuotes(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("GetStringBetweenQuotes(%q) = %q, want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestNormalizeJSONKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "lowercase no special chars",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "uppercase",
			input:    "HELLO",
			expected: "hello",
		},
		{
			name:     "mixed case",
			input:    "HeLLo",
			expected: "hello",
		},
		{
			name:     "with underscores",
			input:    "hello_world",
			expected: "helloworld",
		},
		{
			name:     "with dashes",
			input:    "hello-world",
			expected: "helloworld",
		},
		{
			name:     "with spaces",
			input:    "hello world",
			expected: "helloworld",
		},
		{
			name:     "mixed special chars",
			input:    "Hello_World-With Spaces",
			expected: "helloworldwithspaces",
		},
		{
			name:     "with numbers",
			input:    "hello123_world",
			expected: "hello123world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dirtyjson.NormalizeJSONKey(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeJSONKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
