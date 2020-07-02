package parser

import (
	"encoding/xml"
	"net/http"
)

type opml struct {
	XMLName xml.Name `xml:"opml"`
	Body    body     `xml:"body"`
}

type body struct {
	XMLName  xml.Name  `xml:"body"`
	Outlines []outline `xml:"outline"`
}

type outline struct {
	XMLName xml.Name `xml:"outline"`
	Type    string   `xml:"type,attr"`
	Text    string   `xml:"text,attr"`
	URL     string   `xml:"xmlUrl,attr"`
}

// OpmlItem is an single feed from opml file
type OpmlItem struct {
	Title string
	URL   string
}

// ReadOmpl read stream for opml file structure and parse feeds from file
func ReadOmpl(url string) ([]OpmlItem, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []OpmlItem{}, err
	}
	defer resp.Body.Close()

	var fl opml
	err = xml.NewDecoder(resp.Body).Decode(&fl)
	if err != nil {
		return []OpmlItem{}, err
	}

	if fl.Body.Outlines == nil {
		return []OpmlItem{}, nil
	}

	var result []OpmlItem
	for _, outline := range fl.Body.Outlines {
		if outline.Type == "rss" {
			result = append(result, OpmlItem{Title: outline.Text, URL: outline.URL})
		}
	}

	return result, nil
}
