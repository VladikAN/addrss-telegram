package parser

import (
	"encoding/xml"
	"io"
	"net/http"
	"strings"

	"github.com/vladikan/addrss-telegram/youtube"
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
	XMLName  xml.Name  `xml:"outline"`
	Type     string    `xml:"type,attr"`
	Text     string    `xml:"text,attr"`
	Title    string    `xml:"title,attr"`
	URL      string    `xml:"xmlUrl,attr"`
	Outlines []outline `xml:"outline"`
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

	return decode(resp.Body)
}

func decode(io io.Reader) ([]OpmlItem, error) {
	var fl opml
	err := xml.NewDecoder(io).Decode(&fl)
	if err != nil {
		return []OpmlItem{}, err
	}

	if fl.Body.Outlines == nil {
		return []OpmlItem{}, nil
	}

	return collectOutlines(fl.Body.Outlines), nil
}

func collectOutlines(outlines []outline) []OpmlItem {
	var result []OpmlItem
	for _, item := range outlines {
		if isImportableOutline(item) {
			title := strings.TrimSpace(item.Text)
			if title == "" {
				title = strings.TrimSpace(item.Title)
			}
			result = append(result, OpmlItem{Title: title, URL: strings.TrimSpace(item.URL)})
		}
		if len(item.Outlines) > 0 {
			result = append(result, collectOutlines(item.Outlines)...)
		}
	}
	return result
}

func isImportableOutline(item outline) bool {
	url := strings.TrimSpace(item.URL)
	if url == "" {
		return false
	}
	if youtube.IsYouTubeURL(url) {
		return true
	}
	t := strings.ToLower(strings.TrimSpace(item.Type))
	return t == "rss" || t == "atom"
}
