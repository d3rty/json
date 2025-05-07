package formgen

import (
	"fmt"
	"reflect"
	"strconv"
)

// Introspect walks your config struct and builds a tree of sections,
// each with its fields (including Disabled) and any nested subsections.
func Introspect(cfg interface{}) (*FormModel, error) {
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	model := &FormModel{}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		fi := t.Field(i)
		if fi.Name == "Section" {
			continue
		}
		// toml tag or fallback to field name
		tag := fi.Tag.Get("toml")
		if tag == "" {
			tag = fi.Name
		}

		fv := v.Field(i)
		if fv.Kind() != reflect.Ptr || fv.IsNil() {
			continue
		}

		// the actual struct value for this section
		sectVal := fv.Elem()
		sect := &FormSection{Title: tag}

		// 1) add the embedded Disabled checkbox
		if secEmbed := sectVal.FieldByName("Section"); secEmbed.IsValid() {
			disabled := secEmbed.FieldByName("Disabled").Bool()
			sect.Fields = append(sect.Fields, FormField{
				Label: "Disabled",
				Name:  tag + ".Disabled",
				Value: strconv.FormatBool(disabled),
				Type:  FieldCheckbox,
			})
		}

		// 2) add all the non‑nested leaf fields
		for j := 0; j < sectVal.NumField(); j++ {
			subFi := sectVal.Type().Field(j)
			// skip only the embedded Section (we already grabbed its Disabled)
			if subFi.Name == "Section" {
				continue
			}
			// skip nested structs here (we’ll handle them below)
			if subFi.Type.Kind() == reflect.Ptr && subFi.Type.Elem().Kind() == reflect.Struct {
				continue
			}
			ff := makeFormField(tag, sectVal.Field(j), subFi)
			sect.Fields = append(sect.Fields, ff)
		}

		// 3) now build any nested subsections
		for j := 0; j < sectVal.NumField(); j++ {
			subFi := sectVal.Type().Field(j)
			if subFi.Type.Kind() == reflect.Ptr && subFi.Type.Elem().Kind() == reflect.Struct {
				if sectVal.Field(j).IsNil() {
					continue
				}

				// subsection name
				subTag := subFi.Tag.Get("toml")
				if subTag == "" {
					subTag = subFi.Name
				}

				childVal := sectVal.Field(j).Elem()
				child := &FormSection{Title: subTag}

				// include this subsection’s Disabled
				if secEmbed := childVal.FieldByName("Section"); secEmbed.IsValid() {
					disabled := secEmbed.FieldByName("Disabled").Bool()
					child.Fields = append(child.Fields, FormField{
						Label: "Disabled",
						Name:  tag + "." + subTag + ".Disabled",
						Value: strconv.FormatBool(disabled),
						Type:  FieldCheckbox,
					})
				}

				// now the leaf fields of the subsection
				for k := 0; k < childVal.NumField(); k++ {
					leafFi := childVal.Type().Field(k)
					if leafFi.Name == "Section" {
						continue
					}
					ff := makeFormField(tag+"."+subTag, childVal.Field(k), leafFi)
					child.Fields = append(child.Fields, ff)
				}

				sect.Subsections = append(sect.Subsections, child)
			}
		}

		model.Sections = append(model.Sections, sect)
	}

	return model, nil
}

// helper that builds a FormField from a reflect.Value + StructField
func makeFormField(prefix string, val reflect.Value, fi reflect.StructField) FormField {
	tomlTag := fi.Tag.Get("toml")
	if tomlTag == "" {
		tomlTag = fi.Name
	}
	// field-name
	name := prefix + "." + tomlTag

	// if prefix is only a top‑level section, show "Section → Field",
	// otherwise (nested) show only the field name.
	label := tomlTag

	var ftype FieldType
	var sval string

	switch val.Kind() {
	case reflect.Bool:
		ftype = FieldCheckbox
		sval = strconv.FormatBool(val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ftype = FieldNumber
		sval = fmt.Sprintf("%d", val.Int())
	default:
		ftype = FieldText
		sval = fmt.Sprint(val.Interface())
	}

	// --- ENUM SPECIAL‑CASE ---
	// Detect our BoolFromNumberAlg type and turn it into a <select>
	if fi.Type.Name() == "BoolFromNumberAlg" {
		ftype = FieldSelect
		// map of (value, label). Tweak labels as you like:
		enumDefs := []struct{ Val, Label string }{
			{"0", "Undefined"},
			{"1", "Binary"},
			{"2", "PositiveNegative"},
			{"4", "SignOfOne"},
		}
		opts := make([]Option, len(enumDefs))
		for i, e := range enumDefs {
			opts[i] = Option{Value: e.Val, Label: e.Label}
		}
		return FormField{
			Label:   label,
			Name:    name,
			Value:   sval,
			Type:    ftype,
			Options: opts,
		}
	}

	return FormField{
		Label: label,
		Name:  name,
		Value: sval,
		Type:  ftype,
	}
}
