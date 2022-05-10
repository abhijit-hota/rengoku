package utils

import (
	"bytes"
	"text/template"
)

func Format(tpl string, values map[string]string) string {
	b := &bytes.Buffer{}
	err := template.Must(template.New("").Parse(tpl)).Execute(b, values)
	Must(err)
	return b.String()
}
