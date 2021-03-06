package server

import (
	"fmt"
	"testing"
)

func TestNormalize(t *testing.T) {
	exp := "1-test.com-normalize"
	rst := normalize("1 TEST.com &+ normalize")
	if rst != exp {
		t.Errorf("Expected '%s', but was '%s'", exp, rst)
	}
}

func TestNormalize_LeadingSpace(t *testing.T) {
	exp := "test"
	rst := normalize("  test   ")
	if exp != rst {
		t.Errorf("Expected '%s', but was '%s'", exp, rst)
	}
}

func TestNormalize_SpacesAfterReplace(t *testing.T) {
	exp := "test-test"
	rst := normalize("test++ ++ ++test")
	if exp != rst {
		t.Errorf("Expected '%s', but was '%s'", exp, rst)
	}
}

func TestUnicode(t *testing.T) {
	exp := "1-тест.ком-normalize"
	rst := normalize("1 ТЕСТ.ком &+ normalize")
	if rst != exp {
		t.Errorf("Expected '%s', but was '%s'", exp, rst)
	}
}

func TestSpliURINonUri(t *testing.T) {
	rst := splitURI("baduri")
	if len(rst) != 0 {
		t.Errorf("Expected an empty array, but was %d length", len(rst))
	}
}

func TestSplitURIBySingle(t *testing.T) {
	in := "http://example.com/test.rss"
	rst := splitURI(in)

	if len(rst) != 1 && rst[0] != in {
		t.Errorf("Expected of length 1 and has '%s', but was '%s'", in, rst[0])
	}
}

func TestSplitURIByMany(t *testing.T) {
	in1 := "http://example.com/test1.rss"
	in2 := "http://example.com/test2.rss"
	rst := splitURI(fmt.Sprintf("%s  %s", in1, in2))

	if len(rst) != 2 {
		t.Errorf("Expected of length 2, but was %d", len(rst))
	}

	if rst[0] != in1 {
		t.Errorf("[0] Expected \"%s\", but got \"%s\"", in1, rst[0])
	}
	if rst[1] != in2 {
		t.Errorf("[1] Expected \"%s\", but got \"%s\"", in2, rst[1])
	}
}

func TestSplitWithEmpty(t *testing.T) {
	rst := splitNonEmpty("   ")
	if len(rst) != 0 {
		t.Errorf("Expected empty array, but was %d length", len(rst))
	}
}

func TestSplitNonEmpty(t *testing.T) {
	rst := splitNonEmpty("1   2")
	if len(rst) != 2 {
		t.Errorf("Expected of length 2, but was %d length", len(rst))
	}

	if rst[0] != "1" {
		t.Errorf("[0] Expected \"1\", but got \"%s\"", rst[0])
	}
	if rst[1] != "2" {
		t.Errorf("[1] Expected \"2\", but got \"%s\"", rst[1])
	}
}
