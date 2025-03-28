package dirtytesting_test

import (
	"encoding/json"
	"os"
	"testing"

	dirty "github.com/d3rty/json"
	"github.com/d3rty/json/internal/config"
	. "github.com/d3rty/json/internal/dirtytesting"
	testmodels "github.com/d3rty/json/tests/models"
	"github.com/stretchr/testify/require"
)

func TestGenerateDirtyJSON(t *testing.T) {
	cleanJSONPath := "../../tests/testdata/1.clean.json"
	cleanContents, err := os.ReadFile(cleanJSONPath)
	require.NoError(t, err)
	cleanContents = minifyJSON(t, cleanContents)

	var cleanData testmodels.Item
	err = json.Unmarshal(cleanContents, &cleanData)
	require.NoError(t, err)

	var dcfg DirtifyCfg
	dirtyContents, err := Dirtify[testmodels.Item](cleanContents, &dcfg)
	require.NoError(t, err)

	config.SetGlobal(func(cfg *config.Config) {
		*cfg = *dcfg.Config()
	})

	var recoveredData testmodels.Item
	err = dirty.Unmarshal(dirtyContents, &recoveredData)
	require.NoError(t, err, "failed with config "+dcfg.Config().String()+" on "+string(dirtyContents))

	cleanMap := StructToMap(cleanData)
	recoveredMap := StructToMap(recoveredData)

	require.Equal(t, cleanMap, recoveredMap, "failed with config "+dcfg.Config().String()+" on "+string(dirtyContents))
}

func minifyJSON(t *testing.T, raw []byte) []byte {
	// minify JSON via marshalling round trip
	// here we assume json is valid but just in case we assert errors
	var d map[string]any
	err := json.Unmarshal(raw, &d)
	require.NoError(t, err)

	result, err := json.Marshal(d)
	require.NoError(t, err)
	return result
}
