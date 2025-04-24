package config_test

import (
	"encoding/json"
	"testing"

	"github.com/d3rty/json/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example section struct implementing the disabler interface.

type SectionFoo struct {
	config.Section

	Foo string
}
type SectionBar struct {
	config.Section

	Bar string
	Baz *SectionBarBaz
}

type SectionBarBaz struct {
	config.Section

	BarBaz string
}

type TestConfig struct {
	Foo *SectionFoo
	Bar *SectionBar
}

func TestHandleDefaultFieldDisabled(t *testing.T) {
	cfg := TestConfig{}

	config.HandleDefaultFieldDisabled(&cfg)

	assert.NotNil(t, cfg.Foo)
	assert.NotNil(t, cfg.Bar)
	assert.True(t, cfg.Foo.IsDisabled())
	assert.True(t, cfg.Bar.IsDisabled())
	assert.True(t, cfg.Bar.Baz.IsDisabled())
}

func TestClone(t *testing.T) {
	// Create an original config
	cfg := config.Global()

	// Clone the original config
	cloned := cfg.Clone()

	// in JSON representation they now must be the same as well
	origBytes, err := json.Marshal(cfg)
	require.NoError(t, err)
	clonedBytes, err := json.Marshal(cloned)
	require.NoError(t, err)

	assert.Equal(t, string(origBytes), string(clonedBytes))
}
