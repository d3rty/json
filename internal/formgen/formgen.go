package formgen

// Package formgen provides functionality to generate a JSON schema
// for rendering a configuration form based on your Go `Config` struct.
//
// We’ve moved away from HTML templates to a JSON-first approach.
// The API is:
//   Introspect(cfg any) (*FormModel, error)
// and the core types are:
//   FormModel, FormSection, FormField, Option, FieldType

// Introspect builds a FormModel by reflecting over your Config.
// See internal/formgen/introspect.go for the implementation.
