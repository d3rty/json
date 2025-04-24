package testmodels

import (
	"encoding/json"

	dirty "github.com/d3rty/json"
)

// Option represents a key/value pair in the option array.
type Option struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Options []Option

// Info represents the information block in details.
type Info struct {
	Category string   `json:"category"`
	Rating   int      `json:"rating"`
	Features []string `json:"features"`
	Options  Options  `json:"options"`
}

// Details represents the nested details structure.
type Details struct {
	Description string  `json:"description"`
	Score       float64 `json:"score"`
	WasVerified bool    `json:"was_verified"`
	Info        Info    `json:"info"`
}

// Item is the top-level object in the JSON document.
// It embeds json.Enabled to allow dirty unmarshalling.
type Item struct {
	dirty.Enabled

	ID       int      `json:"id"`
	Name     string   `json:"name"`
	IsActive bool     `json:"is_active"`
	Details  Details  `json:"details"`
	Tags     []string `json:"tags"`
}

// String returns JSON representation of the Item (Clean JSON).
func (i Item) String() string {
	jj, _ := json.Marshal(i)
	return string(jj)
}

// Dirty returns the dirty schema variant for the Item.
// This is used when clean unmarshalling fails so that flexible (dirty)
// parsing can be applied to problematic fields.
func (i Item) Dirty() any {
	return &struct {
		ID       dirty.Number `json:"id"`
		IsActive dirty.Bool   `json:"is_active"`
		Details  struct {
			Score       dirty.Number `json:"score"`
			WasVerified dirty.Bool   `json:"was_verified"`
			Info        struct {
				Rating  dirty.Number `json:"rating"`
				Options []struct {
					Key   string `json:"key"`
					Value string `json:"value"` // TODO: change to scalar
				} `json:"options"`
			} `json:"info"`
		} `json:"details"`
	}{}
}

// Document represents an array of Items.
type Document []Item
