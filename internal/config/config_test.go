package config_test

import (
	"testing"

	"github.com/amberpixels/k1/maybe"
	"github.com/d3rty/json/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalConfig_NoPanic(t *testing.T) {
	require.NotPanics(t, func() {
		_ = config.Global()
	})
}

func TestGlobalConfig_Valid(t *testing.T) {
	globalConfig := config.Global()

	assert.NotNil(t, globalConfig)

	boolConfig := globalConfig.Bool
	assert.NotNil(t, boolConfig)
	assert.False(t, boolConfig.IsDisabled())
	assert.Equal(t, boolConfig.FallbackValue, maybe.False())

	assert.NotNil(t, boolConfig.FromStrings)
	assert.False(t, boolConfig.FromStrings.IsDisabled())
	assert.Equal(t, []string{"true", "yes", "on"}, boolConfig.FromStrings.CustomListForTrue)
	assert.Equal(t, []string{"false", "no", "off", ""}, boolConfig.FromStrings.CustomListForFalse)
	assert.True(t, boolConfig.FromStrings.CaseInsensitive)
	assert.True(t, boolConfig.FromStrings.RespectFromNumbersLogic)

	assert.NotNil(t, boolConfig.FromNumbers)
	assert.False(t, boolConfig.FromNumbers.IsDisabled())
	assert.Equal(t, config.BoolFromNumberBinary, boolConfig.FromNumbers.CustomParseFunc)

	assert.NotNil(t, boolConfig.FromNull)
	assert.False(t, boolConfig.FromNull.IsDisabled())
	assert.False(t, boolConfig.FromNull.Inverse)

	numberConfig := globalConfig.Number

	assert.NotNil(t, numberConfig)
	assert.False(t, numberConfig.IsDisabled())

	assert.NotNil(t, numberConfig.FromStrings)
	assert.False(t, numberConfig.FromStrings.IsDisabled())
	assert.True(t, numberConfig.FromStrings.SpacingAllowed)
	assert.True(t, numberConfig.FromStrings.ExponentNotationAllowed)
	assert.True(t, numberConfig.FromStrings.CommasAllowed)
	assert.Equal(t, config.RoundingAlgFloor, numberConfig.FromStrings.RoundingAlgorithm)

	assert.NotNil(t, numberConfig.FromBools)
	assert.False(t, numberConfig.FromBools.IsDisabled())

	assert.NotNil(t, numberConfig.FromNull)
	assert.False(t, numberConfig.FromNull.IsDisabled())

	dateConfig := globalConfig.Date

	assert.NotNil(t, dateConfig)
	assert.False(t, dateConfig.IsDisabled())
	assert.NotNil(t, dateConfig.Timezone)
	assert.Equal(t, "UTC", dateConfig.Timezone.Default)
	assert.Equal(t, []string{"timezone", "tz"}, dateConfig.Timezone.Fields)

	assert.NotNil(t, dateConfig.FromStrings)
	assert.False(t, dateConfig.FromStrings.IsDisabled())
	assert.True(t, dateConfig.FromStrings.RespectFromNumbersLogic)
	assert.Equal(t, []string{"3:04PM", "15:04", "15:04:05"}, dateConfig.FromStrings.Layouts.Time)
	assert.Equal(t, []string{
		"2006-01-02",
		"2006/01/02",
		"02 Jan 06",
		"02-Jan-06",
		"Mon, 02 Jan 06",
		"Mon, 02-Jan-06",
		"Monday, 02 Jan 06",
		"Monday, 02-Jan-06",
		"02 Jan 2006",
		"02-Jan-2006",
		"Mon, 02 Jan 2006",
		"Mon, 02-Jan-2006",
		"Monday, 02 Jan 2006",
		"Monday, 02-Jan-2006",
	}, dateConfig.FromStrings.Layouts.Date)
	assert.Equal(t, []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05.999999999Z07:00",
	}, dateConfig.FromStrings.Layouts.DateTime)
	assert.True(t, dateConfig.FromStrings.Aliases)

	assert.NotNil(t, dateConfig.FromNumbers)
	assert.False(t, dateConfig.FromNumbers.IsDisabled())
	assert.True(t, dateConfig.FromNumbers.UnixTimestamp)
	assert.True(t, dateConfig.FromNumbers.UnixMilliTimestamp)

	assert.NotNil(t, dateConfig.FromNull)
	assert.False(t, dateConfig.FromNull.IsDisabled())

	flexKeysConfig := globalConfig.FlexKeys
	assert.NotNil(t, flexKeysConfig)
	assert.True(t, flexKeysConfig.IsDisabled())
}
