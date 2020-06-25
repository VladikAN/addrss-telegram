package templates

import (
	"bytes"
	"fmt"
	"text/template"
)

// ToText prints template content
func ToText(name string) (string, error) {
	return ToTextW(name, nil)
}

// ToTextW prints template content with the data
func ToTextW(name string, data interface{}) (string, error) {
	tmpl, err := template.ParseFiles(fmt.Sprintf("templates/en/%s.txt", name))
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, data)
	if err != nil {
		return "", err
	}

	return tpl.String(), nil
}
