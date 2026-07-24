<p align="center">
  <img src="logo.svg" alt="d3rty/json" width="338">
</p>

<div align="center">

### Gracefully handling schema variations.

Flexible JSON unmarshalling for Go, forgiving of dirty real-world data.

[![Go Reference](https://pkg.go.dev/badge/github.com/d3rty/json.svg)](https://pkg.go.dev/github.com/d3rty/json)
[![CI](https://github.com/d3rty/json/actions/workflows/go.yml/badge.svg)](https://github.com/d3rty/json/actions/workflows/go.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/d3rty/json)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-yellow.svg)](LICENSE)

</div>

---

**d3rty/json** parses JSON from sources you can't fully control or trust: third-party APIs with schema flakiness, legacy feeds, AI-generated output. Stringified numbers, `"on"`/`"off"` booleans, drifting key casing - instead of failing on the first mismatch, it decodes strictly first and falls back to a forgiving dirty schema you define per model.

> [!NOTE]
> Work in progress: the API may still change before v1.

```go
data := []byte(`{"id": "123"}`)

var plain struct{ ID int `json:"id"` }
json.Unmarshal(data, &plain)  // error: cannot unmarshal string into Go struct field .id of type int

var event Event               // same shape, dirty-enabled (see Quick Start)
dirty.Unmarshal(data, &event) // nil; event.ID == 123
```

## Install

```bash
go get github.com/d3rty/json
```

```go
import dirty "github.com/d3rty/json"
```

## Quick Start

Embed `dirty.Enabled` into your clean model and describe what may be dirty in a `Dirty()` counterpart. Everything else stays strict.

```go
package main

import (
	"fmt"
	"log"

	dirty "github.com/d3rty/json"
)

// Event is a regular "clean" model with strict types.
type Event struct {
	dirty.Enabled // opts this model into dirty unmarshalling

	Name string `json:"name"`

	// Covered by the dirty schema below: stringified numbers,
	// sloppy booleans, etc. are forgiven.
	ID       int  `json:"id"`
	IsActive bool `json:"is_active"`

	// Not covered by the dirty schema: must be a real bool or it's lost.
	MustBool bool `json:"must_bool"`
}

// Dirty links the clean model to its dirty counterpart.
func (e *Event) Dirty() any {
	return &struct {
		ID       dirty.Number `json:"id"`
		IsActive dirty.Bool   `json:"is_active"`
	}{}
}

func main() {
	data := []byte(`{"id": "123", "name": "Sample Event", "is_active": "on", "must_bool": true}`)

	var event Event
	if err := dirty.Unmarshal(data, &event); err != nil {
		log.Fatal(err)
	}

	fmt.Println(event.ID)       // 123
	fmt.Println(event.IsActive) // true
	fmt.Println(event.MustBool) // true
}
```

The dirty types are `dirty.Number`, `dirty.String`, `dirty.Bool`, `dirty.Array` and `dirty.Object`. Swap `dirty.Enabled` for `dirty.Disabled` to switch a model back to strict stdlib behavior without touching anything else.

## Flexible Keys

Key naming drifts too: `isActive` today, `IsActive` tomorrow, `is-active` from the mobile team. The `FlexKeys` option (off by default) matches keys case-insensitively and across camelCase/snake_case/kebab-case/PascalCase:

```go
dirty.ConfigSetGlobal(func(cfg *dirty.Config) {
	cfg.FlexKeys.Disabled = false
	cfg.FlexKeys.CaseInsensitive = true
	cfg.FlexKeys.ChameleonCase = true
})

// key arrives as PascalCase, value as a string - both forgiven
data := []byte(`{"IsActive": "on"}`)
```

The case-detection machinery is exposed as its own package, [`cases`](cases/).

## Configuration

Every forgiveness rule is a switch in a TOML config: which strings count as `true`, whether `"1 000 000"` or `"1e6"` are numbers, which date layouts to try, and so on. The full annotated list of defaults lives in [`default.toml`](internal/config/default.toml).

```go
// tweak the global config in place
dirty.ConfigSetGlobal(func(cfg *dirty.Config) {
	cfg.Bool.FromStrings.CustomListForTrue = []string{"true", "yes", "on", "da"}
})

// or load a full config from TOML bytes
cfg := dirty.ConfigFromBytes(tomlBytes)
```

Unmarshalling results are classified by color: 🟢 green (decoded cleanly), 🟡 yellow (decoded via forgiving conversions), 🔴 red (data was lost). Reporting this back to the caller is not implemented yet.

## Playground

The [WASM demo](https://d3rty.github.io/json/) runs the real decoder in your browser: paste dirty JSON, toggle config switches, see what gets forgiven.

## Feedback

d3rty/json is a solo, opinionated project - but if you stumbled upon it and have ideas, questions, or bug reports, an [issue](https://github.com/d3rty/json/issues) is always welcome :)

## License

[MIT](LICENSE) © [d3rty](https://github.com/d3rty)
