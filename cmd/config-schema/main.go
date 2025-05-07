package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/d3rty/json/internal/config"
	"github.com/d3rty/json/internal/formgen"
)

func main() {
	cfg := config.Default()

	model, err := formgen.Introspect(cfg)
	if err != nil {
		log.Fatalf("failed to introspect config: %v", err)
	}

	// Marshal to JSON
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(model); err != nil {
		log.Fatalf("failed to encode form model to JSON: %v", err)
	}
}
