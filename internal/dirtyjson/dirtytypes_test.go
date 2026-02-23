package dirtyjson_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/d3rty/json/internal/config"
	"github.com/d3rty/json/internal/dirtyjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDateTimeUnmarshalJSON tests the DateTime.UnmarshalJSON method.
func TestDateTimeUnmarshalJSON(t *testing.T) {
	// Explicitly set global config to the default config
	// This ensures the test uses the same configuration as the production code
	config.SetGlobal(func(cfg *config.Config) {
		// Reset to default config
		cfg.ResetToDefault()
	})

	// Test cases
	testCases := []struct {
		name     string
		input    string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "RFC3339 format",
			input:    `"2023-01-02T15:04:05Z"`,
			expected: time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "ISO8601 format",
			input:    `"2023-01-02T15:04:05"`,
			expected: time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Unix timestamp as number",
			input:    `1672671845`,
			expected: time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Unix timestamp as string",
			input:    `"1672671845"`,
			expected: time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Unix millisecond timestamp as number",
			input:    `1672671845000`,
			expected: time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Unix millisecond timestamp as string",
			input:    `"1672671845000"`,
			expected: time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:    "Invalid format",
			input:   `"not-a-date"`,
			wantErr: true,
		},
		{
			name:    "Boolean value",
			input:   `true`,
			wantErr: true,
		},
		{
			name:    "Object value",
			input:   `{"date": "2023-01-02"}`,
			wantErr: true,
		},
		{
			name:    "Array value",
			input:   `[2023, 1, 2]`,
			wantErr: true,
		},
		{
			name:     "Null value",
			input:    `null`,
			expected: time.Time{},
			wantErr:  false,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var dt dirtyjson.DateTime
			err := json.Unmarshal([]byte(tc.input), &dt)

			if tc.wantErr {
				assert.Error(t, err, "Expected error for input %v", tc.input)
			} else {
				require.NoError(t, err, "Unexpected error for input %v", tc.input)
				assert.Equal(
					t,
					tc.expected,
					time.Time(dt),
					"Expected %v for input %v, got %v",
					tc.expected,
					tc.input,
					time.Time(dt),
				)
			}
		})
	}
}
