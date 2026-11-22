package pdf

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"backend/src/procedure/adapters/pdf/pdftest"
)

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

	if got := pdftest.Lines(t, d.Bytes())[0][0]; got != "Análisis (II) \\ Ñandú ¿sí? ??" {
		t.Errorf("line = %q", got)
	}
}

func TestTextWrapsBetweenWordsAndInsideOnesTooLong(t *testing.T) {
	d := New("Comprobante", "SIGTA")
	d.Text(strings.Repeat("palabra ", 30) + strings.Repeat("x", 200) + "\nfin")

	lines := pdftest.Lines(t, d.Bytes())[0]
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

	pages := pdftest.Lines(t, d.Bytes())
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

// A backslash at the end must not escape the parenthesis that closes the
// string, or the reader loses the rest of the page.
func TestTextEndingInABackslashKeepsTheNextLines(t *testing.T) {
	d := New("Comprobante", "SIGTA")
	d.Text(`C:\`)
	d.Text(`a \(b\) c\\`)
	d.Text("sigue")

	out := d.Bytes()
	if !bytes.Contains(out, []byte(`(C:\\) Tj`)) {
		t.Errorf("the backslash is not escaped: %q", out)
	}
	if got := pdftest.Lines(t, out)[0]; fmt.Sprint(got[:3]) != fmt.Sprint([]string{`C:\`, `a \(b\) c\\`, "sigue"}) {
		t.Errorf("lines = %q", got)
	}
}

func TestWidthOfMeasuresEachFontAndCountsWhatItDoesNotKnowAsTheWidest(t *testing.T) {
	for _, tc := range []struct {
		text  string
		style style
		want  float64
	}{
		{"m", textStyle, 8.33},
		{"m", labelStyle, 8.89},
		{"Ñandú", textStyle, 7.22 + 5.56 + 5.56 + 5.56 + 5.56},
		{"¿", labelStyle, 6.11},
		{"——", labelStyle, 20},
		{"陈", textStyle, 10},
	} {
		if got := widthOf(tc.text, tc.style); math.Abs(got-tc.want) > 0.001 {
			t.Errorf("widthOf(%q) = %.3f, want %.3f", tc.text, got, tc.want)
		}
	}
}
