package templates

import (
	"bytes"
	"fmt"
	"text/template"
)

var output func(name string, data interface{}) (string, error)

// SetTemplateOutput set standard template generation
func SetTemplateOutput() {
	output = func(name string, data interface{}) (string, error) {
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
}

// SetCustomOutput set custom output generation
func SetCustomOutput(out func(name string, data interface{}) (string, error)) {
	output = out
}

// ToText prints template content
func ToText(name string) (string, error) {
	return output(name, nil)
}

// ToTextW prints template content with the data
func ToTextW(name string, data interface{}) (string, error) {
	return output(name, data)
}
