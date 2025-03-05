package dirtytesting_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/d3rty/json/internal/dirtytesting"
	testmodels "github.com/d3rty/json/tests/models"
	"github.com/stretchr/testify/require"
)

// TODO: make a real test (at least smoke) with its own data, but not with testmodels
func TestGenerateDirtyJSON(t *testing.T) {
	cleanJsonPath := "../../testdata/static/1.clean.json"
	contents, err := os.ReadFile(cleanJsonPath)
	require.NoError(t, err)

	dirtyContents, err := dirtytesting.GenerateDirtyJSON(&testmodels.Item{}, contents, 0.7)
	fmt.Println(err)
	fmt.Println(string(dirtyContents))
}
