package formgen

import (
	"fmt"
	"io/fs"
	"reflect"
	"strconv"

	"github.com/d3rty/json/internal/config"
)

const (
	LabelSection = "Section"
)

// Introspect walks the provided cfg (must be a pointer to your Config) and
// builds a FormModel representing each section, its Disabled flag, fields,
// and nested subsections.
func Introspect(cfg any) (*FormModel, error) {
	// Load TOML comments from the embedded config
	tomlFile, err := fs.ReadFile(config.EmbeddedConfig(), "default.toml")
	if err != nil {
		return nil, fmt.Errorf("failed to read default.toml: %w", err)
	}

	comments, err := ParseTOMLComments(string(tomlFile))
	if err != nil {
		return nil, fmt.Errorf("failed to parse TOML comments: %w", err)
	}

	return introspectWithComments(cfg, comments)
}

// introspectWithComments is the internal version that accepts parsed comments.
func introspectWithComments(cfg any, comments TOMLComments) (*FormModel, error) {
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	model := &FormModel{}
	t := v.Type()
	for i := range t.NumField() {
		fi := t.Field(i)
		// skip embedded Section
		if fi.Name == LabelSection {
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
		if secEmbed := sectVal.FieldByName(LabelSection); secEmbed.IsValid() {
			disabled := secEmbed.FieldByName("Disabled").Bool()
			sect.Fields = append(sect.Fields, FormField{
				Name:        tag + ".Disabled",
				Label:       "Disabled",
				Type:        FieldCheckbox,
				Value:       strconv.FormatBool(disabled),
				Description: comments.GetDescription(tag + ".Disabled"),
			})
		}

		// 2) leaf fields
		for j := range sectVal.NumField() {
			subFi := sectVal.Type().Field(j)
			if subFi.Name == LabelSection {
				continue
			}
			// skip nested structs
			if subFi.Type.Kind() == reflect.Ptr && subFi.Type.Elem().Kind() == reflect.Struct {
				continue
			}
			fv2 := sectVal.Field(j)
			sect.Fields = append(sect.Fields, makeFormField(tag, fv2, subFi, comments))
		}

		// 3) nested subsections
		for j := range sectVal.NumField() {
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
				if secEmbed := childVal.FieldByName(LabelSection); secEmbed.IsValid() {
					disabled := secEmbed.FieldByName("Disabled").Bool()
					childPath := fmt.Sprintf("%s.%s.Disabled", tag, subTag)
					child.Fields = append(child.Fields, FormField{
						Name:        childPath,
						Label:       "Disabled",
						Type:        FieldCheckbox,
						Value:       strconv.FormatBool(disabled),
						Description: comments.GetDescription(childPath),
					})
				}

				// child's leaf fields
				for k := range childVal.NumField() {
					leafFi := childVal.Type().Field(k)
					if leafFi.Name == LabelSection {
						continue
					}
					fv3 := childVal.Field(k)
					child.Fields = append(child.Fields, makeFormField(tag+"."+subTag, fv3, leafFi, comments))
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
func makeFormField(prefix string, val reflect.Value, fi reflect.StructField, comments TOMLComments) FormField {
	tomlTag := fi.Tag.Get("toml")
	if tomlTag == "" {
		tomlTag = fi.Name
	}
	name := prefix + "." + tomlTag
	label := tomlTag
	description := comments.GetDescription(name)

	var ftype FieldType
	var sval string

	switch val.Kind() {
	case reflect.Bool:
		ftype = FieldCheckbox
		sval = strconv.FormatBool(val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ftype = FieldNumber
		sval = strconv.FormatInt(val.Int(), 10)
	default:
		// enum BoolFromNumberAlg → select
		if fi.Type == reflect.TypeFor[config.BoolFromNumberAlg]() {
			ftype = FieldSelect
			var opts []Option
			for _, v := range config.ListAvailableBoolFromNumberAlgs() {
				opts = append(opts, Option{
					Value: strconv.FormatUint(uint64(uint8(v)), 10),
					Label: v.String(),
				})
			}
			return FormField{
				Name:        name,
				Label:       label,
				Type:        ftype,
				Description: description,
				//nolint:errcheck // it's fine here
				Value:   strconv.FormatUint(uint64(uint8(val.Interface().(config.BoolFromNumberAlg))), 10),
				Options: opts,
			}
		}

		// fallback → text
		ftype = FieldText
		sval = fmt.Sprint(val.Interface())
	}

	return FormField{
		Name:        name,
		Label:       label,
		Type:        ftype,
		Value:       sval,
		Description: description,
	}
}
