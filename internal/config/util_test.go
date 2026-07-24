package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/d3rty/json/internal/config"
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

	// The clone must deeply equal the original
	assert.Equal(t, cfg, cloned)
}
