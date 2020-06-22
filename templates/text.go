package templates

import (
	"bytes"
	"fmt"
	"text/template"
)

// ToText prints template content
func ToText(name string) string {
	return ToTextW(name, nil)
}

// ToTextW prints template content with the data
func ToTextW(name string, data interface{}) string {
	var tpl bytes.Buffer
	fl, _ := template.ParseFiles(fmt.Sprintf("templates/en/%s.txt", name))
	fl.Execute(&tpl, data)

	return tpl.String()
}
