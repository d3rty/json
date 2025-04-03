package dirtytesting_test

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"testing"

	dirty "github.com/d3rty/json"
	"github.com/d3rty/json/internal/config"
	. "github.com/d3rty/json/internal/dirtytesting"
	testmodels "github.com/d3rty/json/tests/models"
	"github.com/stretchr/testify/require"
)

const (
	sampleTestEnabled           = true
	sampleTestConfigPath        = "../../tests/testdata/0.config.toml"
	sampleTestDirtyContentsPath = "../../tests/testdata/1.test.dirty.json"
	sampleTestCleanContentsPath = "../../tests/testdata/1.clean.json"
)

func TestSample(t *testing.T) {
	if !sampleTestEnabled {
		t.Skip()
	}

	cfg, err := config.Load(sampleTestConfigPath)
	require.NoError(t, err)

	config.SetGlobal(func(globalConfig *config.Config) {
		*globalConfig = *cfg
	})

	dirtyContents, err := os.ReadFile(sampleTestDirtyContentsPath)
	require.NoError(t, err)

	var result testmodels.Item
	err = dirty.Unmarshal(dirtyContents, &result)
	require.NoError(t, err)
	resultMap := StructToMap(result)

	cleanContents, err := os.ReadFile(sampleTestCleanContentsPath)
	require.NoError(t, err)

	var cleanData testmodels.Item
	err = json.Unmarshal(cleanContents, &cleanData)
	require.NoError(t, err)
	cleanMap := StructToMap(cleanData)

	require.Equal(t, cleanMap, resultMap)
}

func TestGenerateDirtyJSON(t *testing.T) {
	cleanJSONPath := "../../tests/testdata/1.clean.json"
	cleanContents, err := os.ReadFile(cleanJSONPath)
	require.NoError(t, err)
	cleanContents = minifyJSON(t, cleanContents)

	var cleanData testmodels.Item
	err = json.Unmarshal(cleanContents, &cleanData)
	require.NoError(t, err)
	cleanMap := StructToMap(cleanData)

	const n = 1000

	for i := range n {
		iStr := strconv.Itoa(i)

		var dcfg DirtifyCfg
		dirtyContents, err := Dirtify[testmodels.Item](cleanContents, &dcfg)
		require.NoError(t, err)

		config.SetGlobal(func(cfg *config.Config) {
			*cfg = *dcfg.Config()
		})

		var recoveredData testmodels.Item
		err = dirty.Unmarshal(dirtyContents, &recoveredData)
		require.NoError(t, err, iStr+": failed with config "+dcfg.Config().String()+" on "+string(dirtyContents))

		recoveredMap := StructToMap(recoveredData)

		require.Equal(t,
			cleanMap, recoveredMap,
			iStr+": failed with config "+dcfg.Config().String()+" on "+string(dirtyContents),
		)
	}
	log.Println("Done ", n, " attempts.")
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
