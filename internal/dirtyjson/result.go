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

// Result stores the result of unmarshalling (as a filled Dirty model)
// TODO: Whole concept of Result should be reconsidered and improved.
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
// 		color: ColorYellow, // TODO: it should be something like v.color(),
// 		dirty: v.result().(*D),
// 	}

// 	return res
// }
