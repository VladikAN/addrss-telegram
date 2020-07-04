package parser

import (
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

var ruDays = map[string]string{
	"вс": "Sun",
	"пн": "Mon",
	"вт": "Tue",
	"ср": "Wed",
	"чт": "Thu",
	"пт": "Fri",
	"сб": "Sat",
}

var ruMonths = map[string]string{
	"янв": "Jan",
	"фев": "Feb",
	"мар": "Mar",
	"апр": "Apr",
	"май": "May",
	"июн": "Jun",
	"июл": "Jul",
	"авг": "Aug",
	"сен": "Sep",
	"окт": "Oct",
	"ноя": "Nov",
	"дек": "Dec",
}

func parseDate(item *gofeed.Item, lang string) (tm *time.Time) {
	tm = item.PublishedParsed
	if tm == nil {
		tm = item.UpdatedParsed // not all feeds has publish/update values, will ignore these feeds for now
	}

	if tm != nil {
		return
	}

	// Try to get raw time string
	raw := item.Published
	if len(raw) == 0 {
		raw = item.Updated
		if len(raw) == 0 {
			return
		}
	}

	raw = strings.ToLower(raw)
	switch strings.ToLower(lang) {
	case "ru":
		{
			raw = replaceLang(raw, ruDays)
			raw = replaceLang(raw, ruMonths)
			tm = parseLayout(raw)
			return
		}
	}

	return tm
}

func replaceLang(str string, rpl map[string]string) string {
	in := str

	for k, v := range rpl {
		in = strings.Replace(in, k, v, 1)
	}

	return in
}

func parseLayout(str string) *time.Time {
	if ts, err := time.Parse(time.RFC822Z, str); err == nil {
		return &ts
	}
	if ts, err := time.Parse(time.RFC822, str); err == nil {
		return &ts
	}

	if ts, err := time.Parse(time.RFC1123Z, str); err == nil {
		return &ts
	}
	if ts, err := time.Parse(time.RFC1123, str); err == nil {
		return &ts
	}

	return nil
}
