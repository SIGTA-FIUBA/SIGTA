package pdf

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/text/encoding/charmap"
)

var (
	contentStream = regexp.MustCompile(`(?s)>>\nstream\n(.*?)endstream`)
	shownText     = regexp.MustCompile(`\(((?:[^()\\]|\\.)*)\) Tj`)
)

// linesOf returns the lines each page shows, back in UTF-8.
func linesOf(t *testing.T, document []byte) [][]string {
	t.Helper()
	var pages [][]string
	for _, stream := range contentStream.FindAllStringSubmatch(string(document), -1) {
		var lines []string
		for _, m := range shownText.FindAllStringSubmatch(stream[1], -1) {
			raw := strings.NewReplacer(`\\`, `\`, `\(`, `(`, `\)`, `)`).Replace(m[1])
			line, err := charmap.Windows1252.NewDecoder().String(raw)
			if err != nil {
				t.Fatal(err)
			}
			lines = append(lines, line)
		}
		pages = append(pages, lines)
	}
	return pages
}

func TestBytesIsAPDFWhoseTableFindsEveryObject(t *testing.T) {
	d := New("Comprobante", "SIGTA")
	d.Text("hola")
	out := d.Bytes()

	if !bytes.HasPrefix(out, []byte("%PDF-1.4\n")) || !bytes.HasSuffix(out, []byte("%%EOF\n")) {
		t.Fatalf("not a PDF: %q", out)
	}
	at := bytes.LastIndex(out, []byte("startxref\n"))
	xref, err := strconv.Atoi(strings.Fields(string(out[at:]))[1])
	if err != nil || !bytes.HasPrefix(out[xref:], []byte("xref\n")) {
		t.Fatalf("startxref does not point to the table: %v", err)
	}
	entries := strings.Split(string(out[xref:]), "\n")[3:]
	for i := 1; ; i++ {
		fields := strings.Fields(entries[i-1])
		if len(fields) != 3 {
			if i != 8 { // catalog, pages, two fonts, info, one page and its content
				t.Errorf("the table has %d objects, want 7", i-1)
			}
			break
		}
		offset, err := strconv.Atoi(fields[0])
		if err != nil || !bytes.HasPrefix(out[offset:], fmt.Appendf(nil, "%d 0 obj\n", i)) {
			t.Errorf("object %d is not at %s", i, fields[0])
		}
	}
	length := regexp.MustCompile(`/Length (\d+) >>\nstream\n`).FindSubmatchIndex(out)
	n, _ := strconv.Atoi(string(out[length[2]:length[3]]))
	if !bytes.HasPrefix(out[length[1]+n:], []byte("endstream")) {
		t.Errorf("the length of the content is not %d", n)
	}
}

func TestTextKeepsSpanishAndMarksWhatWinAnsiLacks(t *testing.T) {
	d := New("Comprobante", "SIGTA")
	d.Text("Análisis (II) \\ Ñandú ¿sí? 日本")

	if got := linesOf(t, d.Bytes())[0][0]; got != "Análisis (II) \\ Ñandú ¿sí? ??" {
		t.Errorf("line = %q", got)
	}
}

func TestTextWrapsBetweenWordsAndInsideOnesTooLong(t *testing.T) {
	d := New("Comprobante", "SIGTA")
	d.Text(strings.Repeat("palabra ", 30) + strings.Repeat("x", 200) + "\nfin")

	lines := linesOf(t, d.Bytes())[0]
	lines = lines[:len(lines)-1] // the footer
	if len(lines) < 4 || lines[len(lines)-1] != "fin" {
		t.Fatalf("lines = %q", lines)
	}
	for _, line := range lines {
		if w := widthOf(line, textStyle); w > lineWidth {
			t.Errorf("%q is %.0f wide, more than %.0f", line, w, lineWidth)
		}
	}
	if !strings.HasPrefix(lines[0], "palabra palabra") || strings.HasSuffix(lines[0], " ") {
		t.Errorf("first line = %q, want words that fit", lines[0])
	}
}

func TestDocumentGoesOnToANewPageAndNumbersThem(t *testing.T) {
	d := New("Comprobante", "SIGTA · Trámite P-1")
	for i := range 80 {
		d.Field("Paso", fmt.Sprintf("línea %d", i))
	}

	pages := linesOf(t, d.Bytes())
	if len(pages) != 2 {
		t.Fatalf("%d pages, want 2", len(pages))
	}
	for i, page := range pages {
		if want := fmt.Sprintf("SIGTA · Trámite P-1 · Página %d de 2", i+1); page[len(page)-1] != want {
			t.Errorf("footer of page %d = %q, want %q", i+1, page[len(page)-1], want)
		}
	}
	if !strings.Contains(strings.Join(pages[1], "\n"), "línea 79") {
		t.Errorf("the second page does not go on: %q", pages[1])
	}
}
