package parser

import (
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func TestParseDate_RU(t *testing.T) {
	ex, _ := time.Parse("2006-01-02T15:0400 +0300", "2020-07-04T15:09:00 +0300")
	tm := "Сб, 4 июл 2020 15:09:00 +0300"
	rt := parseDate(&gofeed.Item{Published: tm}, "RU")

	if rt == nil {
		t.Errorf("Unable to parse input string '%s'", tm)
	}

	if *rt != ex {
		t.Errorf("Expected '%s', but was '%s'", ex, *rt)
	}
}
