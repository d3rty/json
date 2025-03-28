package config_test

import (
	"testing"

	"github.com/d3rty/json/internal/config"
	"github.com/stretchr/testify/assert"
)

// Example section struct implementidng the disabler interface.
type MySection struct {
	Disabled bool
}
type SectionFoo struct {
	MySection

	Foo string
}
type SectionBar struct {
	MySection

	Bar string
}

func (s *MySection) IsDisabled() bool {
	return s.Disabled
}

type TestConfig struct {
	Foo *SectionFoo
	Bar *SectionBar
}

func TestSetDefaults(t *testing.T) {
	cfg := TestConfig{}

	config.SetDefaults(&cfg)

	assert.NotNil(t, cfg.Foo)
	assert.NotNil(t, cfg.Bar)
	assert.True(t, cfg.Foo.IsDisabled())
	assert.True(t, cfg.Bar.IsDisabled())
}
