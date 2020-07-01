package templates

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

var langs []string = []string{"en", "ru"}
var output func(lang string, name string, data interface{}) (string, error)

// SetTemplateOutput set standard template generation
func SetTemplateOutput() {
	output = func(lang string, name string, data interface{}) (string, error) {
		loc := parseLang(lang)
		tmpl, err := template.ParseFiles(fmt.Sprintf("templates/%s/%s.txt", loc, name))
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
func SetCustomOutput(out func(lang string, name string, data interface{}) (string, error)) {
	output = out
}

// ToText prints template content
func ToText(lang string, name string) (string, error) {
	return output(lang, name, nil)
}

// ToTextW prints template content with the data
func ToTextW(lang string, name string, data interface{}) (string, error) {
	return output(lang, name, data)
}

func parseLang(lang string) string {
	loc := strings.ToLower(lang)
	for _, name := range langs {
		if name == loc {
			return loc
		}
	}

	return "en"
}
