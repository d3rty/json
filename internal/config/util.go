package config

import (
	"reflect"

	"github.com/BurntSushi/toml"
)

// clone via toml round-trip.
// It's a simple (but not the most efficient) way to clone the config.
// Config is safe for marshalling (that's by design): It will never contain functions, etc.
// We can live with this solution until we need increase performance.
// TODO(1): refactor so it's not a round-trip via []byte
// TODO(2): remove panics for live prod code.
func clone(cfg *Config) *Config {
	contents, err := toml.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	var clone Config
	if err := toml.Unmarshal(contents, &clone); err != nil {
		panic(err)
	}
	return &clone
}

const (
	fieldNameDisabled = "Disabled"
	fieldNameSection  = "Section"
)

type disabler interface {
	IsDisabled() bool
}

// disablerType is the reflect.Type of the disabler interface.
var disablerType = reflect.TypeFor[disabler]()

// handleDefaultFieldDisabled sets default value for `Disabled bool`
// So if a section (TOML Table) is not presented, it will be changed to an empty table with Disabled=true
// So, the IsDisabled() call is possible and returns true.
func handleDefaultFieldDisabled(v any) {
	if v == nil {
		return
	}

	// Ensure we have a pointer to a struct.
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return
	}
	rv = rv.Elem()
	rt := rv.Type()

	// Loop through each field.
	for i := range rv.NumField() {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		if field.Kind() == reflect.Struct {
			handleDefaultFieldDisabled(field.Addr().Interface())
			continue
		}
		if field.Kind() != reflect.Ptr {
			continue
		}
		if field.Type().Elem().Kind() != reflect.Struct {
			continue
		}

		if !field.IsNil() {
			handleDefaultFieldDisabled(field.Interface())
			continue
		}

		if fieldType.Type.Implements(disablerType) || fieldType.Type.Elem().Implements(disablerType) {
			newInstance := reflect.New(fieldType.Type.Elem())
			csField := newInstance.Elem().FieldByName(fieldNameSection)
			if csField.IsValid() && csField.CanSet() && csField.Kind() == reflect.Struct {
				disabledField := csField.FieldByName(fieldNameDisabled)
				if disabledField.IsValid() && disabledField.CanSet() && disabledField.Kind() == reflect.Bool {
					disabledField.SetBool(true)
				}
			}
			field.Set(newInstance)
			handleDefaultFieldDisabled(newInstance.Interface())
		}
	}
}
