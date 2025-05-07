package main

import (
	"fmt"
	"log"

	"github.com/d3rty/json/internal/config"
	"github.com/d3rty/json/internal/formgen"
)

func main() {
	// 1) load your default config
	cfg := config.Default()

	// 2) introspect into a form model
	model, err := formgen.Introspect(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// 3) render HTML
	htmlForm, err := formgen.Render(model)
	if err != nil {
		log.Fatal(err)
	}

	// 4) e.g. write it to stdout or serve via HTTP
	fmt.Println(htmlForm)
}
