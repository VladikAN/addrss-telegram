package server

import (
	"regexp"
	"strings"
)

func normalize(in string) string {
	rg, _ := regexp.Compile("[^\\p{L}\\d_ \\.]+")
	nrm := rg.ReplaceAllString(in, "")

	rg, _ = regexp.Compile("\\s+")
	nrm = rg.ReplaceAllString(nrm, " ")

	nrm = strings.TrimSpace(nrm)
	nrm = strings.ReplaceAll(nrm, " ", "-")
	nrm = strings.ToLower(nrm)

	return nrm
}

func splitURI(in string) []string {
	raw := splitNonEmpty(in)
	rg, _ := regexp.Compile("http(s)?://[\\w\\d\\.\\-]+/[\\w\\d\\.\\-\\?\\&\\/\\=]+")

	var rst []string
	for _, n := range raw {
		if !rg.MatchString(n) {
			continue
		}

		rst = append(rst, n)
	}

	return rst
}

func splitNonEmpty(in string) []string {
	if len(in) == 0 {
		return []string{}
	}

	var rst []string
	raw := strings.Split(in, " ")
	for _, n := range raw {
		if len(n) == 0 {
			continue
		}

		rst = append(rst, n)
	}

	return rst
}
