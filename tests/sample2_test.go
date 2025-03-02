package tests

import (
	"encoding/json"
	"testing"

	dirty "github.com/d3rty/json"
	"github.com/d3rty/json/internal/config"
	testmodels "github.com/d3rty/json/tests/models"
	"github.com/stretchr/testify/require"
)

func TestSample2_Clean(t *testing.T) {

	contents := ReadSampleFile(t, "static/2.clean")

	// Ensure std unmarshal works for the read file
	var stdResult testmodels.Document
	require.NoError(t,
		json.Unmarshal(contents, &stdResult),
	)

	// 1. Set config to minimal (reset)
	config.UpdateGlobal(config.Reset)
	var clean1Result testmodels.Document
	require.NoError(t,
		dirty.Unmarshal(contents, &clean1Result),
	)
	require.Equal(t, stdResult, clean1Result)

	// 2. Set default (dirty) config
	config.UpdateGlobal(config.Default)
	var clean2Result testmodels.Document
	require.NoError(t,
		dirty.Unmarshal(contents, &clean2Result),
	)

	require.Equal(t, stdResult, clean2Result)
}

// TODO: Dirty (2) should be read and compared to clean (2)
