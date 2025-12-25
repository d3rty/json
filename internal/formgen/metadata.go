package formgen

// FieldType represents the kind of input widget to render.
type FieldType string

const (
	FieldText     FieldType = "text"
	FieldNumber   FieldType = "number"
	FieldCheckbox FieldType = "checkbox"
	FieldSelect   FieldType = "select"
)

// Option represents one <option> in a <select> box.
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// FormField describes a single input field in the form.
type FormField struct {
	Name        string    `json:"name"`  // e.g. "Bool.FromStrings.CustomListForTrue"
	Label       string    `json:"label"` // e.g. "CustomListForTrue"
	Type        FieldType `json:"type"`
	Description string    `json:"description,omitempty"` // tooltip/help text from TOML comments

	Value   string   `json:"value"`             // the value as string
	Options []Option `json:"options,omitempty"` // the value only for selects
}

// FormSection represents a group of fields (and nested subsections).
type FormSection struct {
	Title       string         `json:"title"`
	Fields      []FormField    `json:"fields"`
	Subsections []*FormSection `json:"subsections,omitempty"`
}

// FormModel is the root JSON schema for the form.
type FormModel struct {
	Sections []*FormSection `json:"sections"`
}
