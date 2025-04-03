package dirtytests

import (
	"encoding/json"
	"testing"

	dirty "github.com/d3rty/json"
	testmodels "github.com/d3rty/json/tests/models"
	"github.com/stretchr/testify/require"
)

func TestSample1_Clean(t *testing.T) {
	contents := ReadSampleFile(t, "1.clean")

	// Ensure std unmarshal works for the read file
	var stdResult testmodels.Item
	require.NoError(t,
		json.Unmarshal(contents, &stdResult),
	)

	// 1. Set config to minimal (reset)
	dirty.ConfigSetGlobal(func(cfg *dirty.Config) {
		cfg.ResetToEmpty()
	})
	var clean1Result testmodels.Item
	require.NoError(t,
		dirty.Unmarshal(contents, &clean1Result),
	)
	require.Equal(t, stdResult, clean1Result)

	// 2. Set default (dirty) config
	dirty.ConfigSetGlobal(func(cfg *dirty.Config) {
		cfg.ResetToDefault()
	})

	var clean2Result testmodels.Item
	require.NoError(t,
		dirty.Unmarshal(contents, &clean2Result),
	)

	require.Equal(t, stdResult, clean2Result)
}

func TestSample1_Dirty_Yellow(t *testing.T) {
	contents := ReadSampleFile(t, "1.dirty-yellow")

	// std should fail as types don't match
	var stdResult testmodels.Item
	require.Error(t,
		json.Unmarshal(contents, &stdResult),
	)

	// Set config to minimal (reset)
	// as it should behave as std - it should fail as well
	dirty.ConfigSetGlobal(func(cfg *dirty.Config) {
		cfg.ResetToEmpty()
	})
	var dirty1Result testmodels.Item
	require.Error(t,
		dirty.Unmarshal(contents, &dirty1Result),
	)

	// Default config should work although.
	dirty.ConfigSetGlobal(func(cfg *dirty.Config) {
		cfg.ResetToDefault()
	})
	var dirt2yResult testmodels.Item
	require.NoError(t,
		dirty.Unmarshal(contents, &dirt2yResult),
	)

	require.JSONEq(t,
		`{
			"id":1,
			"name":"Item 1",
			"is_active":true,
			"details":{
				"description":"Description for item 1",
				"score": 9.5,
				"was_verified": false,
				"info":{
					"category":"Category A",
					"rating":4,
					"features":["fast","reliable"],
					"options":[{"key":"priority","value":"high"},{"key":"limit","value":"10"}]
				}
			},
			"tags":["alpha","beta"]
		}`, dirt2yResult.String(),
	)
}

func TestSample1_Dirty_YellowChameleon(t *testing.T) {
	t.Skip("TODO fix the test")
	contents := ReadSampleFile(t, "1.dirty-yellow.keys")

	// std should fail as types don't match
	var stdResult testmodels.Item
	require.Error(t,
		json.Unmarshal(contents, &stdResult),
	)

	// Set config to minimal (reset)
	// as it should behave as std - it should fail as well
	dirty.ConfigSetGlobal(func(cfg *dirty.Config) {
		cfg.ResetToEmpty()
	})
	var dirty1Result testmodels.Item
	require.Error(t,
		dirty.Unmarshal(contents, &dirty1Result),
	)

	// Maximum config: Default + FlexKeys
	dirty.ConfigSetGlobal(func(cfg *dirty.Config) {
		cfg.ResetToDefault()
		cfg.FlexKeys.Disabled = false
		cfg.FlexKeys.ChameleonCase = true
		cfg.FlexKeys.CaseInsensitive = true
	})

	var dirt2yResult testmodels.Item
	require.NoError(t,
		dirty.Unmarshal(contents, &dirt2yResult),
	)

	require.JSONEq(t,
		`{
			"id":1,
			"name":"Item 1",
			"is_active":true,
			"details":{
				"description":"Description for item 1",
				"score": 9.5,
				"was_verified": false,
				"info":{
					"category":"Category A",
					"rating":4,
					"features":["fast","reliable"],
					"options":[{"key":"priority","value":"high"},{"key":"limit","value":10}]
				}
			},
			"tags":["alpha","beta"]
		}`, dirt2yResult.String(),
	)
}
