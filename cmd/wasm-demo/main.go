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

// cleanJSON is exposed to JavaScript. It takes two arguments:
//
//	args[0] = raw JSON string
//	args[1] = TOML config string (optional, from generateTOML(getConfigFromForm()))
//
// It unmarshals into Event (with dirty fallback), then pretty‑prints or returns an error.
func cleanJSON(this js.Value, args []js.Value) any {
	console := js.Global().Get("console")

	if len(args) < 1 {
		return js.ValueOf("error: cleanJSON(jsonStr, configToml) expects at least one argument")
	}
	jsonStr := args[0].String()

	console.Call("log", "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	console.Call("log", "📥 INPUT:")
	console.Call("log", jsonStr)

	// Apply config from TOML if provided
	if len(args) >= 2 && !args[1].IsUndefined() && !args[1].IsNull() {
		configToml := args[1].String()
		console.Call("log", "⚙️  CONFIG:")
		console.Call("log", configToml)

		if cfg := dirty.ConfigFromBytes([]byte(configToml)); cfg != nil {
			dirty.ConfigSetGlobal(func(c *dirty.Config) {
				*c = *cfg
			})
		}
	}

	var e Event
	if err := dirty.Unmarshal([]byte(jsonStr), &e); err != nil {
		errMsg := "unmarshal error: " + err.Error()
		console.Call("log", "❌ ERROR:", errMsg)
		console.Call("log", "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		return js.ValueOf(errMsg)
	}

	out, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		errMsg := "marshal error: " + err.Error()
		console.Call("log", "❌ ERROR:", errMsg)
		console.Call("log", "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		return js.ValueOf(errMsg)
	}

	console.Call("log", "✅ RESULT:")
	console.Call("log", string(out))
	console.Call("log", "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return js.ValueOf(string(out))
}

func main() {
	// Expose cleanJSON to the JS global scope.
	js.Global().Set("cleanJSON", js.FuncOf(cleanJSON))

	// Block forever so the WASM module stays alive.
	select {}
}
