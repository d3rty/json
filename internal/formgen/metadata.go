package formgen

type FieldType string

const (
	FieldText     FieldType = "text"
	FieldNumber   FieldType = "number"
	FieldCheckbox FieldType = "checkbox"
	FieldSelect   FieldType = "select"
)

type Option struct {
	Value string // the <option value="...">
	Label string // the human label inside the <option>
}

type FormField struct {
	Label   string    // human‑friendly
	Name    string    // e.g. "Bool.FallbackValue"
	Value   string    // default/current value
	Type    FieldType // how to render
	Options []Option  // for selects (e.g. enums)
}

type FormSection struct {
	Title       string // e.g. "Bool" or "FromStrings"
	Fields      []FormField
	Subsections []*FormSection // nested <fieldset>s
}

type FormModel struct {
	Sections []*FormSection
}
