package formgen

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/d3rty/json/internal/config"
)

// Introspect walks the provided cfg (must be a pointer to your Config) and
// builds a FormModel representing each section, its Disabled flag, fields,
// and nested subsections.
func Introspect(cfg interface{}) (*FormModel, error) {
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	model := &FormModel{}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		fi := t.Field(i)
		// skip embedded Section
		if fi.Name == "Section" {
			continue
		}

		tag := fi.Tag.Get("toml")
		if tag == "" {
			tag = fi.Name
		}

		fv := v.Field(i)
		if fv.Kind() != reflect.Ptr || fv.IsNil() {
			continue
		}

		sectVal := fv.Elem()
		sect := &FormSection{Title: tag}

		// 1) embedded Disabled flag
		if secEmbed := sectVal.FieldByName("Section"); secEmbed.IsValid() {
			disabled := secEmbed.FieldByName("Disabled").Bool()
			sect.Fields = append(sect.Fields, FormField{
				Name:  tag + ".Disabled",
				Label: "Disabled",
				Type:  FieldCheckbox,
				Value: strconv.FormatBool(disabled),
			})
		}

		// 2) leaf fields
		for j := 0; j < sectVal.NumField(); j++ {
			subFi := sectVal.Type().Field(j)
			if subFi.Name == "Section" {
				continue
			}
			// skip nested structs
			if subFi.Type.Kind() == reflect.Ptr && subFi.Type.Elem().Kind() == reflect.Struct {
				continue
			}
			fv2 := sectVal.Field(j)
			sect.Fields = append(sect.Fields, makeFormField(tag, fv2, subFi))
		}

		// 3) nested subsections
		for j := 0; j < sectVal.NumField(); j++ {
			subFi := sectVal.Type().Field(j)
			if subFi.Type.Kind() == reflect.Ptr && subFi.Type.Elem().Kind() == reflect.Struct {
				if sectVal.Field(j).IsNil() {
					continue
				}

				subTag := subFi.Tag.Get("toml")
				if subTag == "" {
					subTag = subFi.Name
				}

				childVal := sectVal.Field(j).Elem()
				child := &FormSection{Title: subTag}

				// child's Disabled
				if secEmbed := childVal.FieldByName("Section"); secEmbed.IsValid() {
					disabled := secEmbed.FieldByName("Disabled").Bool()
					child.Fields = append(child.Fields, FormField{
						Name:  fmt.Sprintf("%s.%s.Disabled", tag, subTag),
						Label: "Disabled",
						Type:  FieldCheckbox,
						Value: strconv.FormatBool(disabled),
					})
				}

				// child's leaf fields
				for k := 0; k < childVal.NumField(); k++ {
					leafFi := childVal.Type().Field(k)
					if leafFi.Name == "Section" {
						continue
					}
					fv3 := childVal.Field(k)
					child.Fields = append(child.Fields, makeFormField(tag+"."+subTag, fv3, leafFi))
				}

				sect.Subsections = append(sect.Subsections, child)
			}
		}

		model.Sections = append(model.Sections, sect)
	}

	return model, nil
}

// makeFormField creates a FormField for a single struct field, including
// special handling for your BoolFromNumberAlg enum (pulling from config.All...)
func makeFormField(prefix string, val reflect.Value, fi reflect.StructField) FormField {
	tomlTag := fi.Tag.Get("toml")
	if tomlTag == "" {
		tomlTag = fi.Name
	}
	name := prefix + "." + tomlTag
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
		// enum BoolFromNumberAlg → select
		if fi.Type == reflect.TypeFor[config.BoolFromNumberAlg]() {
			ftype = FieldSelect
			var opts []Option
			for _, v := range config.ListAvailableBoolFromNumberAlgs() {
				opts = append(opts, Option{
					Value: fmt.Sprint(uint8(v)),
					Label: v.String(),
				})
			}
			return FormField{Name: name, Label: label, Type: ftype,
				Value:   fmt.Sprint(uint8(val.Interface().(config.BoolFromNumberAlg))),
				Options: opts,
			}
		}

		// fallback → text
		ftype = FieldText
		sval = fmt.Sprint(val.Interface())
	}

	return FormField{Name: name, Label: label, Type: ftype, Value: sval}
}
