package dirty

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// decodeValue decodes JSON tokens into the provided reflect.Value.
func decodeValue(dec *json.Decoder, rv reflect.Value) error {
	// Ensure rv is a pointer.
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("v must be a non-nil pointer")
	}

	// Dereference to the actual value.
	val := rv.Elem()
	return decodeInto(dec, val)
}

// decodeInto recursively decodes JSON from dec into the provided value.
func decodeInto(dec *json.Decoder, val reflect.Value) error {
	switch val.Kind() {
	case reflect.Struct:
		// Check if the struct (via its address) implements CustomUnmarshaler.
		if val.CanAddr() {
			original := val.Addr().Interface()
			if dd, ok := original.(d3rtyDecoder); ok {
				// We expect our magic struct to also be Dirtyable.
				dirtyable, ok := original.(Dirtyable)
				if !ok {
					return errors.New("magic struct does not implement Dirtyable")
				}
				schema := dirtyable.Dirty() // schema is a pointer to a relaxed version
				// Read the entire JSON object for this struct into a raw buffer.
				var raw json.RawMessage
				if err := dec.Decode(&raw); err != nil {
					return fmt.Errorf("buffering raw JSON failed: %w", err)
				}
				// First, try decoding the raw JSON into the original (clean) type.
				cleanErr := json.Unmarshal(raw, original)
				if cleanErr == nil {
					// Clean decode worked; nothing more to do.
					return nil
				}
				// If clean decoding fails, initialize dirty schema.
				if container, ok := original.(d3rtyContainer); ok {
					container.init(schema)
				}
				// Now create a new decoder from the same raw JSON bytes.
				dirtyDec := json.NewDecoder(bytes.NewReader(raw))
				if err := dd.decode(dirtyDec); err != nil {
					return fmt.Errorf("dirty decode failed: %w", err)
				}
				// Merge: marshal the dirty schema to JSON...
				buf, err := json.Marshal(schema)
				if err != nil {
					return fmt.Errorf("marshalling dirty schema failed: %w", err)
				}
				// ... then unmarshal into the original struct.
				if err := json.Unmarshal(buf, original); err != nil {
					return fmt.Errorf("merging dirty schema failed: %w", err)
				}
				return nil
			}
		}
		return decodeStruct(dec, val)

	case reflect.Slice:
		return decodeSlice(dec, val)

	case reflect.Array:
		return decodeArray(dec, val)

	case reflect.Map:
		return decodeMap(dec, val)

	default:
		// For scalars or any type that cannot contain nested structs,
		// just use standard decoding.
		ptr := reflect.New(val.Type())
		if err := dec.Decode(ptr.Interface()); err != nil {
			return err
		}
		val.Set(ptr.Elem())
		return nil
	}
}

// decodeStruct decodes a JSON object into a struct value.
func decodeStruct(dec *json.Decoder, val reflect.Value) error {
	// Expect the next token to be '{'
	t, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok || delim != '{' {
		return errors.New("expected start of object '{'")
	}

	typ := val.Type()
	// Loop over all fields in the JSON object.
	for dec.More() {
		// Field name (as string).
		t, err := dec.Token()
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
				if err := decodeInto(dec, reflect.Indirect(fv)); err != nil {
					return fmt.Errorf("error decoding field %q: %w", sf.Name, err)
				}
				break
			}
		}
		// Field not found in struct: skip the value.
		if !fieldFound {
			var skip interface{}
			if err := dec.Decode(&skip); err != nil {
				return err
			}
		}
	}

	// Consume the ending '}' token.
	t, err = dec.Token()
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
func decodeSlice(dec *json.Decoder, val reflect.Value) error {
	// Expect '[' as the starting token.
	t, err := dec.Token()
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
	for dec.More() {
		// Create a new element value.
		elem := reflect.New(elemType).Elem()
		if err := decodeInto(dec, elem); err != nil {
			return fmt.Errorf("error decoding slice element: %w", err)
		}
		sliceVal = reflect.Append(sliceVal, elem)
	}

	// Consume the ending ']' token.
	t, err = dec.Token()
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
func decodeArray(dec *json.Decoder, val reflect.Value) error {
	// Expect '[' as the starting token.
	t, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok || delim != '[' {
		return errors.New("expected start of array '['")
	}

	length := val.Len()
	for i := 0; i < length; i++ {
		if !dec.More() {
			return errors.New("not enough elements in JSON array")
		}
		elem := val.Index(i)
		if err := decodeInto(dec, elem); err != nil {
			return fmt.Errorf("error decoding array element %d: %w", i, err)
		}
	}
	// If there are extra elements, skip them.
	for dec.More() {
		var skip interface{}
		if err := dec.Decode(&skip); err != nil {
			return err
		}
	}
	// Consume the ending ']' token.
	t, err = dec.Token()
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
func decodeMap(dec *json.Decoder, val reflect.Value) error {
	// If the map is nil, allocate it.
	if val.IsNil() {
		val.Set(reflect.MakeMap(val.Type()))
	}

	// Expect '{' as the start.
	t, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok || delim != '{' {
		return errors.New("expected start of object '{' for map")
	}

	keyType := val.Type().Key()
	elemType := val.Type().Elem()
	for dec.More() {
		// Map keys are provided as tokens (usually strings).
		t, err := dec.Token()
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
		if err := decodeInto(dec, elem); err != nil {
			return err
		}
		val.SetMapIndex(key, elem)
	}
	// Consume the ending '}' token.
	t, err = dec.Token()
	if err != nil {
		return err
	}
	delim, ok = t.(json.Delim)
	if !ok || delim != '}' {
		return errors.New("expected end of object '}' for map")
	}
	return nil
}
