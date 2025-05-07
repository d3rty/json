package formgen

import (
	"bytes"
	"html/template"
)

var formTmpl = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Config Form</title>
</head>
<body>
<form method="POST" action="/config/save">
{{- range .Sections}}
  {{ template "section" . }}
{{- end}}
  <button type="submit">Save</button>
</form>
</body>
</html>

{{ define "section" }}
<fieldset>
  <legend>{{ .Title }}</legend>

  {{- range .Fields}}
    {{- $f := . }}
  <div>
    <label for="{{ $f.Name }}">{{ $f.Label }}</label>

    {{- if eq $f.Type "` + string(FieldCheckbox) + `" }}
      <input type="checkbox" id="{{ $f.Name }}" name="{{ $f.Name }}"{{ if eq $f.Value "true" }} checked{{ end }}>
    {{- else if eq $f.Type "` + string(FieldNumber) + `" }}
      <input type="number" id="{{ $f.Name }}" name="{{ $f.Name }}" value="{{ $f.Value }}">
    {{- else if eq $f.Type "` + string(FieldSelect) + `" }}
      <select id="{{ $f.Name }}" name="{{ $f.Name }}">
        {{- range $opt := $f.Options }}
          <option value="{{ $opt.Value }}"{{ if eq $f.Value $opt.Value }} selected{{ end }}>{{ $opt.Label }}</option>
        {{- end}}
      </select>
    {{- else }}
      <input type="text" id="{{ $f.Name }}" name="{{ $f.Name }}" value="{{ $f.Value }}">
    {{- end}}

  </div>
  {{- end}}

  {{- range .Subsections}}
    {{ template "section" . }}
  {{- end}}
</fieldset>
{{ end }}
`

// Render takes a FormModel and returns the generated HTML.
func Render(model *FormModel) (string, error) {
	t := template.Must(template.New("cfgForm").Parse(formTmpl))
	var buf bytes.Buffer
	if err := t.Execute(&buf, model); err != nil {
		return "", err
	}
	return buf.String(), nil
}
