package parser

import (
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func TestParseDate_RU(t *testing.T) {
	ex, _ := time.Parse(time.RFC1123Z, "Sat, 04 Jul 2020 15:09:00 +0300")
	tm := "Сб, 4 июл 2020 15:09:00 +0300"
	rt := parseDate(&gofeed.Item{Published: tm}, "ru-RU")

	if rt == nil {
		t.Errorf("Unable to parse input string '%s'", tm)
	}

	if *rt != ex {
		t.Errorf("Expected '%s', but was '%s'", ex, *rt)
	}
}
