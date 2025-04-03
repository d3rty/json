package dirtytests

import (
	"testing"

	dirty "github.com/d3rty/json"
	"github.com/d3rty/json/internal/config"
	testmodels "github.com/d3rty/json/tests/models"
	"github.com/stretchr/testify/require"
)

func XTestSampleSandbox(t *testing.T) {
	contents := []byte(`{
    "ID": true,
    "TAGS": ["alpha", "beta"],
    "details": {
      "DESCRIPTION": "Description for item 1",
      "info": {
        "CATEGORY": "Category A",
        "FEATURES": ["fast", "reliable"],
        "RATING": 4,
        "options": [{ "key": "priority", "value": "high" }]
      },
      "score": 9.5,
      "was_verified": false
    },
    "is_active": true,
    "name": "Item 1"
  }`)

	config.SetGlobal(func(cfg *config.Config) {
		*cfg = *config.FromBytes([]byte(`{
    "Bool": {
      "Allowed": false,
      "FromStrings": {
        "Allowed": false,
        "CustomListForTrue": null,
        "CustomListForFalse": null,
        "CaseInsensitive": false,
        "FalseForEmptyString": false,
        "RespectFromNumbersLogic": false,
        "FallbackValue": null
      },
      "FromNumbers": {
        "Allowed": false,
        "CustomParseFunc": "",
        "FallbackValue": null
      },
      "FromNull": { "Allowed": false, "Inverse": false }
    },
    "Number": {
      "Allowed": true,
      "FromStrings": {
        "Allowed": false,
        "SpacingAllowed": false,
        "ExponentNotationAllowed": false,
        "CommasAllowed": false,
        "FloatishAllowed": false
      },
      "FromBools": { "Allowed": true },
      "FromNull": { "Allowed": false }
    },
    "FlexKeys": {
      "Allowed": true,
      "CaseInsensitive": true,
      "ChameleonCase": true
    }
  }`))
	})

	var dirtyResult testmodels.Item
	require.Error(t,
		dirty.Unmarshal(contents, &dirtyResult),
	)

	// TODO
	_ = dirtyResult
	// fmt.Println("=> ", dirtyResult)
}
