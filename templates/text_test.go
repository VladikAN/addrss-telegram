package templates

import "testing"

func TestParseLang_WithKnownLang(t *testing.T) {
	exp := "ru"
	r := parseLang("RU")
	if r != exp {
		t.Errorf("Expected '%s', but was '%s'", exp, r)
	}
}

func TestParseLang_WithUnknownLang(t *testing.T) {
	exp := "en"
	r := parseLang("unknown")
	if r != exp {
		t.Errorf("Expected '%s', but was '%s'", exp, r)
	}
}
