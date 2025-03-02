package tests

import (
	"encoding/json"
	"testing"

	dirty "github.com/d3rty/json"
	"github.com/d3rty/json/internal/config"
	testmodels "github.com/d3rty/json/tests/models"
	"github.com/stretchr/testify/require"
)

func TestSample1_Clean(t *testing.T) {

	contents := ReadSampleFile(t, "static/1.clean")

	// Ensure std unmarshal works for the read file
	var stdResult testmodels.Item
	require.NoError(t,
		json.Unmarshal(contents, &stdResult),
	)

	// 1. Set config to minimal (reset)
	config.UpdateGlobal(config.Reset)
	var clean1Result testmodels.Item
	require.NoError(t,
		dirty.Unmarshal(contents, &clean1Result),
	)
	require.Equal(t, stdResult, clean1Result)

	// 2. Set default (dirty) config
	config.UpdateGlobal(config.Default)
	var clean2Result testmodels.Item
	require.NoError(t,
		dirty.Unmarshal(contents, &clean2Result),
	)

	require.Equal(t, stdResult, clean2Result)
}

func TestSample1_Dirty_Yellow(t *testing.T) {
	contents := ReadSampleFile(t, "static/1.dirty-yellow")

	// std should fail as types don't match
	var stdResult testmodels.Item
	require.Error(t,
		json.Unmarshal(contents, &stdResult),
	)

	// Set config to minimal (reset)
	// as it should behave as std - it should fail as well
	config.UpdateGlobal(config.Reset)
	var dirty1Result testmodels.Item
	require.Error(t,
		dirty.Unmarshal(contents, &dirty1Result),
	)

	// Default config should work although.
	config.UpdateGlobal(config.Default)
	var dirt2yResult testmodels.Item
	require.NoError(t,
		dirty.Unmarshal(contents, &dirt2yResult),
	)

	require.JSONEq(t,
		dirt2yResult.String(),
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
		}`,
	)
}

func TestSample1_Dirty_YellowChameleon(t *testing.T) {
	contents := ReadSampleFile(t, "static/1.dirty-yellow.keys")

	// std should fail as types don't match
	var stdResult testmodels.Item
	require.Error(t,
		json.Unmarshal(contents, &stdResult),
	)

	// Set config to minimal (reset)
	// as it should behave as std - it should fail as well
	config.UpdateGlobal(config.Reset)
	var dirty1Result testmodels.Item
	require.Error(t,
		dirty.Unmarshal(contents, &dirty1Result),
	)

	// Maximum config: Default + FlexKeys
	config.UpdateGlobal(config.Default, func(cfg *config.Config) {
		cfg.FlexKeys.Allowed = true
		cfg.FlexKeys.ChameleonCase = true
		cfg.FlexKeys.CaseInsensitive = true
	})

	var dirt2yResult testmodels.Item
	require.NoError(t,
		dirty.Unmarshal(contents, &dirt2yResult),
	)

	require.JSONEq(t,
		dirt2yResult.String(),
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
		}`,
	)
}
