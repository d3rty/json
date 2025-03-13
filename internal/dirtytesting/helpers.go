package dirtytesting

import (
	"encoding/json"
)

// StructToMap converts any struct to a map[string]any via JSON round-trip.
func StructToMap(s any) map[string]any {
	var m map[string]any
	b, err := json.Marshal(s)
	if err != nil {
		panic("structToMap: failed to marshal struct " + err.Error())
	}
	if err := json.Unmarshal(b, &m); err != nil {
		panic("structToMap: failed to unmarshal struct " + err.Error())
	}
	return m
}
