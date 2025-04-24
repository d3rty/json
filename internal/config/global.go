package config

import (
	"sync"
)

// globalConfig is the package-level variable storing the config.
//
//nolint:gochecknoglobals // it's an OK case for global variable
var (
	globalConfig = defaultConfig()
	mu           sync.RWMutex
)

// Global returns a copy of the global configuration.
// The returned copy is a clone. Modifying of it doesn't affect the original config.
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
