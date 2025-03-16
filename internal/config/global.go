package config

import (
	"sync"
)

// globalConfig is the package-level variable storing the config.
var (
	//nolint:gochecknoglobals // decide if we need this linter?
	globalConfig *Config
	//nolint:gochecknoglobals // decide if we need this linter?
	mu sync.RWMutex
)

//nolint:gochecknoinits // decide if we need this linter?
func init() {
	globalConfig = defaultConfig()
}

// Global returns a copy of the global configuration.
// Returned copy is a clone. It's modifying doesn't affect the original config.
func Global() *Config {
	mu.RLock()
	defer mu.RUnlock()

	return clone(globalConfig)
}

// SetGlobal updates the global configuration.
func SetGlobal(updateFns ...func(config *Config)) {
	if len(updateFns) == 0 {
		return
	}

	mu.Lock()
	for _, updateFn := range updateFns {
		updateFn(globalConfig)
	}
	mu.Unlock()
}
