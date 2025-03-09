package dirtytesting_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	dirty "github.com/d3rty/json"
	"github.com/d3rty/json/internal/dirtytesting"
	testmodels "github.com/d3rty/json/tests/models"
	"github.com/stretchr/testify/require"
)

// TODO: make a real test (at least smoke) with its own data, but not with testmodels
// WIP: now we go, and we can debug things
func TestGenerateDirtyJSON(t *testing.T) {
	cleanJsonPath := "../../testdata/static/1.clean.json"
	cleanContents, err := os.ReadFile(cleanJsonPath)
	require.NoError(t, err)

	var cleanData testmodels.Item
	err = json.Unmarshal(cleanContents, &cleanData)
	require.NoError(t, err)

	dirtyContents, err := dirtytesting.GenerateDirtyJSON(&testmodels.Item{}, cleanContents, 0.7)
	fmt.Println(err)
	fmt.Println(string(dirtyContents))

	var recovered testmodels.Item
	err = dirty.Unmarshal(dirtyContents, &recovered)
	require.NoError(t, err)
	require.Equal(t, cleanData, recovered)
}
