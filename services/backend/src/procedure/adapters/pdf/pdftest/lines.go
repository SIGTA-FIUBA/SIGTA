// Package pdftest reads back what a PDF of package pdf shows, for the tests
// of the documents it writes.
package pdftest

import (
	"regexp"
	"testing"

	"golang.org/x/text/encoding/charmap"
)

var (
	contentStream = regexp.MustCompile(`(?s)>>\nstream\n(.*?)endstream`)
	shownText     = regexp.MustCompile(`\(((?:[^()\\]|\\.)*)\) Tj`)
	escaped       = regexp.MustCompile(`\\(.)`)
)

// Lines returns the lines each page shows, in UTF-8, the footer last.
func Lines(t testing.TB, document []byte) [][]string {
	t.Helper()
	var pages [][]string
	for _, stream := range contentStream.FindAllStringSubmatch(string(document), -1) {
		var lines []string
		for _, m := range shownText.FindAllStringSubmatch(stream[1], -1) {
			line, err := charmap.Windows1252.NewDecoder().String(escaped.ReplaceAllString(m[1], "$1"))
			if err != nil {
				t.Fatal(err)
			}
			lines = append(lines, line)
		}
		pages = append(pages, lines)
	}
	return pages
}
