package config

// Following exported functions are only accessible from tests.
//
//nolint:gochecknoglobals // `export_test.go` is meant to declare global variables
var (
	HandleDefaultFieldDisabled = handleDefaultFieldDisabled
	Clone                      = clone
)
