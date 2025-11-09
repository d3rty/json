//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"syscall/js"

	dirty "github.com/d3rty/json"
)

// Event is our predefined model.
// Embedding dirty.Enabled opts this struct into the dirty.Unmarshal flow.
type Event struct {
	dirty.Enabled

	Name     string `json:"name"`
	ID       int    `json:"id"`
	IsActive bool   `json:"is_active"`
	MustBool bool   `json:"must_bool"`
}

// Dirty returns the “dirty” variant struct used when the clean decode fails.
// Here we switch out int/bool fields for dirty.Number/dirty.Bool so they can
// accept strings, nulls, etc., if the input is flaky.
func (e *Event) Dirty() any {
	return &struct {
		Name     dirty.String `json:"name"`
		ID       dirty.Number `json:"id"`
		IsActive dirty.Bool   `json:"is_active"`
		MustBool dirty.Bool   `json:"must_bool"`
	}{}
}

// cleanJSON is exposed to JavaScript. It takes one argument:
//
//	args[0] = raw JSON string
//
// It unmarshals into Event (with dirty fallback), then pretty‑prints or returns an error.
func cleanJSON(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf("error: cleanJSON(jsonStr) expects one argument")
	}
	jsonStr := args[0].String()

	var e Event
	if err := dirty.Unmarshal([]byte(jsonStr), &e); err != nil {
		return js.ValueOf("unmarshal error: " + err.Error())
	}

	out, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return js.ValueOf("marshal error: " + err.Error())
	}
	return js.ValueOf(string(out))
}

func main() {
	// Expose cleanJSON to the JS global scope.
	js.Global().Set("cleanJSON", js.FuncOf(cleanJSON))

	// Block forever so the WASM module stays alive.
	select {}
}
