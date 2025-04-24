package dirtyjson

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/d3rty/json/internal/config"
)

// Token is needed for the `CleanDecoder` interface.
type Token = json.Token

// CleanDecoder is an interface that stands for the part of `*json.Decoder` that is required in dirty decoding.
// It also allows us to switch from std `json.Decoder` to other faster (std-compatible) decoders.
type CleanDecoder interface {
	Token() (Token, error)
	More() bool
	Decode(v any) error
}

// Decoder is the dirty decoder, it wraps the "clean" decoder.
type Decoder struct {
	clean CleanDecoder
}

// NewDecoder creates a new dirty decoder.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{clean: json.NewDecoder(r)}
}

func (dec *Decoder) Token() (Token, error) { return dec.clean.Token() }
func (dec *Decoder) More() bool            { return dec.clean.More() }

// Decode decodes the given value respecting dirty schema if possible.
func (dec *Decoder) Decode(val any) error {
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("v must be a non-nil pointer")
	}

	return dec.decode(rv.Elem()) // decode into dereferenced value
}

// cleanDecode is just a wrapper for dec.clean.Decode.
func (dec *Decoder) cleanDecode(val any) error {
	return dec.clean.Decode(val)
}

// The `decode` recursively decodes JSON from dec into the provided value.
// It does a regular decoding routine for all types recursively until it reaches the structs.
// On structs, it tries to decodeDirty if possible.
func (dec *Decoder) decode(val reflect.Value) error {
	switch val.Kind() {
	case reflect.Struct:
		// TODO possible bug here when struct is actually a SmartScalar
		if !val.CanAddr() {
			return dec.decodeStruct(val)
		}
		dirtySchemeOwner, ok := val.Addr().Interface().(Dirtyable)
		if !ok {
			return dec.decodeStruct(val)
		}

		return dec.decodeDirty(dirtySchemeOwner)

	case reflect.Slice:
		return dec.decodeSlice(val)

	case reflect.Array:
		return dec.decodeArray(val)

	case reflect.Map:
		return dec.decodeMap(val)

	default:
		// For scalars and others use clean decoding.
		ptr := reflect.New(val.Type())
		if err := dec.cleanDecode(ptr.Interface()); err != nil {
			return err
		}
		val.Set(ptr.Elem())
		return nil
	}
}

func (dec *Decoder) decodeDirty(v Dirtyable) error {
	// Dirty implementation currently is not so fast.
	// But our aim is the result (dirty decoding) rather than the speed.

	// we need to recover JSON data for the current struct
	// use raw json.Decode for it
	var raw json.RawMessage
	if err := dec.cleanDecode(&raw); err != nil {
		return fmt.Errorf("buffering tee JSON failed: %w", err)
	}

	// thisDec is the decoder for the current struct
	curDec := NewDecoder(bytes.NewReader(raw))

	cfg := getConfig(context.Background())

	// Start with clean decoding: it will fill the clean fields
	// If no FlexKeys is enabled, and decoding was OK, just return here.
	cleanErr := curDec.cleanDecode(v)
	if cleanErr == nil {
		if cfg.FlexKeys.Disabled {
			return nil
		}
	}

	// User may manually disable dirty decoding via embedding Disabled atom
	if _, ok := v.(interface{ isDisabled() }); ok {
		return cleanErr
	}

	// re-init because so we can read again
	curDec = NewDecoder(bytes.NewReader(raw))

	// scheme is a pointer to a dirty version of the struct
	scheme := v.Dirty()

	// If clean decoding fails, initialize dirty schema.
	container, ok := v.(d3rtyContainer)
	if !ok {
		return errors.New("expected dirty container")
	}

	container.init(scheme)
	res := container.result()

	if cfg.FlexKeys.IsDisabled() {
		if err := curDec.cleanDecode(res); err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("dirty decode failed: %w", err)
		}
	} else {
		if err := curDec.Decode(res); err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("dirty decode failed: %w", err)
		}
	}

	// Merge: marshal the dirty schema to JSON...
	// TODO: this part looks the ugliest. We're OK for now, but it's not good.
	buf, err := json.Marshal(scheme)
	if err != nil {
		return fmt.Errorf("marshalling dirty schema failed: %w", err)
	}
	// ... then unmarshal into the original struct.
	if err := json.Unmarshal(buf, v); err != nil {
		return fmt.Errorf("merging dirty schema failed: %w", err)
	}
	return nil
}

// decodeStruct decodes a JSON object into a struct value.
func (dec *Decoder) decodeStruct(val reflect.Value) error {
	// Expect the next token to be '{'
	t, err := dec.clean.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok || delim != '{' {
		return errors.New("expected start of object '{'")
	}

	cfg := getConfig(context.Background())

	typ := val.Type()
	// Loop over all fields in the JSON object.
	for dec.clean.More() {
		// Field name (as string).
		t, err := dec.clean.Token()
		if err != nil {
			return err
		}
		fieldName, ok := t.(string)
		if !ok {
			return errors.New("expected field name to be a string")
		}

		// Find the matching struct field.
		fieldFound := false
		for i := range typ.NumField() {
			sf := typ.Field(i)
			// Only consider exported fields.
			if sf.PkgPath != "" {
				continue
			}

			// Determine the JSON key: check the `json` tag.
			tag := sf.Tag.Get("json")
			name := sf.Name
			if tag != "" && tag != "-" {
				parts := strings.Split(tag, ",")
				if parts[0] != "" {
					name = parts[0]
				}
			}

			if keyMatch(fieldName, name, cfg) {
				fieldFound = true
				fv := val.Field(i)
				// If the field is a pointer and nil, allocate it.
				if fv.Kind() == reflect.Ptr && fv.IsNil() {
					fv.Set(reflect.New(fv.Type().Elem()))
				}
				// Decode into the underlying value.
				if err := dec.decode(reflect.Indirect(fv)); err != nil {
					return fmt.Errorf("error decoding field %q: %w", sf.Name, err)
				}
				break
			}
		}
		// Field not found in struct: skip the value.
		if !fieldFound {
			var skip any
			if err := dec.clean.Decode(&skip); err != nil {
				return err
			}
		}
	}

	// Consume the ending '}' token.
	t, err = dec.clean.Token()
	if err != nil {
		return err
	}
	delim, ok = t.(json.Delim)
	if !ok || delim != '}' {
		return errors.New("expected end of object '}'")
	}
	return nil
}

// keyMatch matches JSON keys (input JSON vs. model) corresponding to the given config.
func keyMatch(jsonKey, modelKey string, cfg *config.Config) bool {
	if jsonKey == modelKey {
		return true
	}
	if cfg.FlexKeys.IsDisabled() {
		return false
	}
	cfgFlexKeys := cfg.FlexKeys

	if !cfgFlexKeys.CaseInsensitive && !cfgFlexKeys.ChameleonCase {
		return false
	}

	if cfgFlexKeys.CaseInsensitive && !cfgFlexKeys.ChameleonCase {
		return strings.EqualFold(jsonKey, modelKey)
	}

	// For ChameleonCase currently we simply normalize keys:
	// we make them lowercase and "one-wordish" (omitting hyphens, underscores and spaces)
	return normalizeJSONKey(jsonKey) == normalizeJSONKey(modelKey)
}

// decodeSlice decodes a JSON array into a slice value.
func (dec *Decoder) decodeSlice(val reflect.Value) error {
	// Expect '[' as the starting token.
	t, err := dec.clean.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok || delim != '[' {
		return errors.New("expected start of array '['")
	}

	elemType := val.Type().Elem()
	sliceVal := reflect.MakeSlice(val.Type(), 0, 0)

	// Process each element.
	for dec.clean.More() {
		// Create a new element value.
		elem := reflect.New(elemType).Elem()
		if err := dec.decode(elem); err != nil {
			return fmt.Errorf("error decoding slice element: %w", err)
		}
		sliceVal = reflect.Append(sliceVal, elem)
	}

	// Consume the ending ']' token.
	t, err = dec.clean.Token()
	if err != nil {
		return err
	}
	delim, ok = t.(json.Delim)
	if !ok || delim != ']' {
		return errors.New("expected end of array ']'")
	}
	val.Set(sliceVal)
	return nil
}

// decodeArray decodes a JSON array into an array value (fixed length).
func (dec *Decoder) decodeArray(val reflect.Value) error {
	// Expect '[' as the starting token.
	t, err := dec.clean.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok || delim != '[' {
		return errors.New("expected start of array '['")
	}

	length := val.Len()
	for i := range length {
		if !dec.clean.More() {
			return errors.New("not enough elements in JSON array")
		}
		elem := val.Index(i)
		if err := dec.decode(elem); err != nil {
			return fmt.Errorf("error decoding array element %d: %w", i, err)
		}
	}
	// If there are extra elements, skip them.
	for dec.clean.More() {
		var skip any
		if err := dec.clean.Decode(&skip); err != nil {
			return err
		}
	}
	// Consume the ending ']' token.
	t, err = dec.clean.Token()
	if err != nil {
		return err
	}
	delim, ok = t.(json.Delim)
	if !ok || delim != ']' {
		return errors.New("expected end of array ']'")
	}
	return nil
}

// decodeMap decodes a JSON object into a map value.
// We assume the map key type is either string or can be decoded from a JSON string.
func (dec *Decoder) decodeMap(val reflect.Value) error {
	// If the map is nil, allocate it.
	if val.IsNil() {
		val.Set(reflect.MakeMap(val.Type()))
	}

	// Expect '{' as the start.
	t, err := dec.clean.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok || delim != '{' {
		return errors.New("expected start of object '{' for map")
	}

	keyType := val.Type().Key()
	elemType := val.Type().Elem()
	for dec.clean.More() {
		// Map keys are provided as tokens (usually strings).
		t, err := dec.clean.Token()
		if err != nil {
			return err
		}
		keyStr, ok := t.(string)
		if !ok {
			return errors.New("expected map key to be a string")
		}

		// Convert keyStr to the key type.
		var key reflect.Value
		if keyType.Kind() == reflect.String {
			key = reflect.ValueOf(keyStr)
		} else {
			// For other key types, decode from a JSON string.
			keyPtr := reflect.New(keyType)
			if err := json.Unmarshal(fmt.Appendf(nil, "%q", keyStr), keyPtr.Interface()); err != nil {
				return fmt.Errorf("error converting map key: %w", err)
			}
			key = keyPtr.Elem()
		}

		// Decode the value.
		elem := reflect.New(elemType).Elem()
		if err := dec.decode(elem); err != nil {
			return err
		}
		val.SetMapIndex(key, elem)
	}
	// Consume the ending '}' token.
	t, err = dec.clean.Token()
	if err != nil {
		return err
	}
	delim, ok = t.(json.Delim)
	if !ok || delim != '}' {
		return errors.New("expected end of object '}' for map")
	}

	return nil
}
