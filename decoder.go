package dirty

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type Token = json.Token

type CleanDecoder interface {
	Token() (Token, error)
	More() bool
	Decode(v any) error
}

type Decoder struct {
	clean CleanDecoder
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{clean: json.NewDecoder(r)}
}

func (dec *Decoder) Decode(val any) error {
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("v must be a non-nil pointer")
	}
	return dec.decodeInto(rv.Elem()) // decode into dereferenced value
}

// decodeInto recursively decodes JSON from dec into the provided value.
func (dec *Decoder) decodeInto(val reflect.Value) error {
	switch val.Kind() {
	case reflect.Struct:
		// Check if the struct (via its address) implements CustomUnmarshaler.
		if val.CanAddr() {
			original := val.Addr().Interface()
			if dirtyable, ok := original.(Dirtyable); ok {
				return dec.decodeDirty(dirtyable)
			}
		}
		return dec.decodeStruct(val)

	case reflect.Slice:
		return dec.decodeSlice(val)

	case reflect.Array:
		return dec.decodeArray(val)

	case reflect.Map:
		return dec.decodeMap(val)

	default:
		// For scalars or any type that cannot contain nested structs,
		// just use standard decoding.
		ptr := reflect.New(val.Type())
		if err := dec.clean.Decode(ptr.Interface()); err != nil {
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
	if err := dec.clean.Decode(&raw); err != nil {
		return fmt.Errorf("buffering tee JSON failed: %w", err)
	}

	// thisDec is the decoder for the current struct
	curDec := NewDecoder(bytes.NewReader(raw))

	// Try clean decoding
	if cleanErr := curDec.clean.Decode(v); cleanErr == nil {
		// Clean decode worked; nothing more to do.
		return nil
	}
	// re-init because so we can read again
	curDec = NewDecoder(bytes.NewReader(raw))

	// scheme is a pointer to a dirty version of the struct
	scheme := v.Dirty()

	// If clean decoding fails, initialize dirty schema.
	container, ok := v.(d3rtyContainer)
	if !ok {
		return fmt.Errorf("expected dirty container")
	}
	container.init(scheme)

	res := container.result()
	if err := curDec.clean.Decode(res); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("dirty decode failed: %w", err)
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
		for i := 0; i < typ.NumField(); i++ {
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
			if name == fieldName {
				fieldFound = true
				fv := val.Field(i)
				// If field is a pointer and nil, allocate it.
				if fv.Kind() == reflect.Ptr && fv.IsNil() {
					fv.Set(reflect.New(fv.Type().Elem()))
				}
				// Decode into the underlying value.
				if err := dec.decodeInto(reflect.Indirect(fv)); err != nil {
					return fmt.Errorf("error decoding field %q: %w", sf.Name, err)
				}
				break
			}
		}
		// Field not found in struct: skip the value.
		if !fieldFound {
			var skip interface{}
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
		if err := dec.decodeInto(elem); err != nil {
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
	for i := 0; i < length; i++ {
		if !dec.clean.More() {
			return errors.New("not enough elements in JSON array")
		}
		elem := val.Index(i)
		if err := dec.decodeInto(elem); err != nil {
			return fmt.Errorf("error decoding array element %d: %w", i, err)
		}
	}
	// If there are extra elements, skip them.
	for dec.clean.More() {
		var skip interface{}
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
			if err := json.Unmarshal([]byte(fmt.Sprintf("%q", keyStr)), keyPtr.Interface()); err != nil {
				return fmt.Errorf("error converting map key: %w", err)
			}
			key = keyPtr.Elem()
		}

		// Decode the value.
		elem := reflect.New(elemType).Elem()
		if err := dec.decodeInto(elem); err != nil {
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
