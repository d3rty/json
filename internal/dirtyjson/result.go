package dirtyjson

// Color means the color of dirty unmarshalling.
type Color int

const (
	// ColorGreen means 100% clean unmarshalling.
	// Data could be unmarshalled directly into the target clean struct.
	ColorGreen Color = iota

	// ColorYellow means successful (lossless) unmarshalling via dirty schema.
	// At least one dirty concept was used. See `Warnings` to list them.
	ColorYellow

	// ColorRed means partially successful (lossy) unmarshalling via dirty schema.
	// At least one field was lost during unmarshalling. See `Errors` to list them.
	ColorRed
)

// TODO(github.com/d3rty/json/issues/14) Concept of colored result (green/yellow/red) to be considered.
//        This file is just a tiny draft (not used anywhere yet)

// Result stores the result of unmarshalling (as a filled Dirty model).
type Result[D any] struct {
	color Color
	dirty *D
}

func (r *Result[D]) Result() *D      { return r.dirty }
func (r *Result[D]) Color() Color    { return r.color }
func (r *Result[D]) Warnings() []any { return nil }
func (r *Result[D]) Errors() []any   { return nil }

// result := dirty.ExtractResult[EventDirty](&e).
// func ExtractResult[D any](v d3rtyContainer) *Result[D] {
// 	res := &Result[D]{
// 		color: ColorYellow,
// 		dirty: v.result().(*D),
// 	}

// 	return res
// }
