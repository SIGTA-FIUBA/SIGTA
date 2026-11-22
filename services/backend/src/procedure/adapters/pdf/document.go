// Package pdf writes simple text documents as PDF: A4 pages of lines in the
// standard Helvetica fonts, which every reader has, so nothing is embedded.
// Text goes in WinAnsiEncoding, which covers Spanish; a character outside it
// is written as "?".
package pdf

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

const (
	pageWidth  = 595.28
	pageHeight = 841.89
	margin     = 56.0
	lineWidth  = pageWidth - 2*margin
)

type style struct {
	font    string
	size    float64
	leading float64
	gray    float64
	widths  *[95]float64
}

var (
	titleStyle   = style{font: "F2", size: 16, leading: 22, widths: &helveticaBoldWidths}
	headingStyle = style{font: "F2", size: 12, leading: 18, widths: &helveticaBoldWidths}
	labelStyle   = style{font: "F2", size: 10, leading: 14, widths: &helveticaBoldWidths}
	textStyle    = style{font: "F1", size: 10, leading: 14, widths: &helveticaWidths}
	noteStyle    = style{font: "F1", size: 8, leading: 11, gray: 0.4, widths: &helveticaWidths}
)

// Document lays out its lines top to bottom and starts a new page when one
// does not fit. Every page ends with the footer and its number.
type Document struct {
	title  string
	footer string
	pages  []*bytes.Buffer
	y      float64
}

// Info is what the document says about itself: the title readers show for it
// and the footer of its pages.
type Info struct {
	Title  string
	Footer string
}

func New(info Info) *Document {
	d := &Document{title: info.Title, footer: info.Footer}
	d.newPage()
	return d
}

func (d *Document) newPage() {
	d.pages = append(d.pages, &bytes.Buffer{})
	d.y = pageHeight - margin
}

func (d *Document) Title(text string)   { d.write(text, titleStyle, 0) }
func (d *Document) Heading(text string) { d.Space(); d.write(text, headingStyle, 0) }
func (d *Document) Text(text string)    { d.write(text, textStyle, 0) }
func (d *Document) Note(text string)    { d.write(text, noteStyle, 0) }

// Indented is text under a line that is not, such as the details of an item.
func (d *Document) Indented(text string) { d.write(text, textStyle, 12) }

// Bold is a line that stands out among the text.
func (d *Document) Bold(text string) { d.write(text, labelStyle, 0) }

// Field writes a label and its value, the value wrapped under itself.
func (d *Document) Field(label, value string) {
	const valueAt = 130.0
	lines := wrap(value, textStyle, lineWidth-valueAt)
	d.fit(textStyle.leading * float64(len(lines)))
	d.show(label, labelStyle, margin)
	for i, line := range lines {
		if i > 0 {
			d.y -= textStyle.leading
			d.fit(textStyle.leading)
		}
		d.show(line, textStyle, margin+valueAt)
	}
	d.y -= textStyle.leading
}

func (d *Document) Space() { d.y -= textStyle.leading / 2 }

func (d *Document) write(text string, s style, indent float64) {
	for _, line := range wrap(text, s, lineWidth-indent) {
		d.fit(s.leading)
		d.show(line, s, margin+indent)
		d.y -= s.leading
	}
}

// fit starts a new page if height does not fit above the footer.
func (d *Document) fit(height float64) {
	if d.y-height < margin+noteStyle.leading {
		d.newPage()
	}
}

func (d *Document) show(line string, s style, x float64) {
	page := d.pages[len(d.pages)-1]
	fmt.Fprintf(page, "BT %.2f g /%s %.1f Tf %.2f %.2f Td (%s) Tj ET\n", s.gray, s.font, s.size, x, d.y-s.size, escape(encode(line)))
}

// Bytes writes the document: the catalog, the pages and their fonts, the
// content of every page and the table that says where each object starts.
func (d *Document) Bytes() []byte {
	var out bytes.Buffer
	var offsets []int
	object := func(body string) {
		offsets = append(offsets, out.Len())
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", len(offsets), body)
	}
	out.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
	// Objects 1 to 5 come first; page i is object 6+2i and its content 7+2i.
	kids := make([]string, len(d.pages))
	for i := range d.pages {
		kids[i] = fmt.Sprintf("%d 0 R", 6+2*i)
	}
	object("<< /Type /Catalog /Pages 2 0 R >>")
	object(fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), len(d.pages)))
	object("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>")
	object("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>")
	object(fmt.Sprintf("<< /Title (%s) /Producer (SIGTA) >>", escape(encode(d.title))))
	for i, page := range d.pages {
		object(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] /Resources << /Font << /F1 3 0 R /F2 4 0 R >> >> /Contents %d 0 R >>", pageWidth, pageHeight, 7+2*i))
		footer := fmt.Sprintf("%s · Página %d de %d", d.footer, i+1, len(d.pages))
		content := page.String() + fmt.Sprintf("BT %.2f g /F1 %.1f Tf %.2f %.2f Td (%s) Tj ET\n", noteStyle.gray, noteStyle.size, margin, margin, escape(encode(footer)))
		object(fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(content), content))
	}
	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(offsets)+1)
	for _, offset := range offsets {
		fmt.Fprintf(&out, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R /Info 5 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets)+1, xref)
	return out.Bytes()
}

// encode turns text into WinAnsiEncoding, which is Windows-1252, with the
// control characters as spaces.
func encode(text string) string {
	out := make([]byte, 0, len(text))
	for _, r := range text {
		b, ok := charmap.Windows1252.EncodeRune(r)
		switch {
		case r < ' ':
			b = ' '
		case !ok:
			b = '?'
		}
		out = append(out, b)
	}
	return string(out)
}

func escape(encoded string) string {
	return strings.NewReplacer(`\`, `\\`, `(`, `\(`, `)`, `\)`).Replace(encoded)
}

// wrap splits text into lines that fit width: at the line breaks it has,
// between words, and inside a word too long for a line of its own.
func wrap(text string, s style, width float64) []string {
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		line := ""
		for _, word := range strings.Fields(paragraph) {
			candidate := word
			if line != "" {
				candidate = line + " " + word
			}
			if widthOf(candidate, s) <= width {
				line = candidate
				continue
			}
			if line != "" {
				lines = append(lines, line)
			}
			line = ""
			for _, r := range word {
				if line != "" && widthOf(line+string(r), s) > width {
					lines = append(lines, line)
					line = ""
				}
				line += string(r)
			}
		}
		lines = append(lines, line)
	}
	return lines
}

// widthOf measures text with the widths of its font, in thousandths of the
// font size. A letter with an accent is as wide as the letter without it, as
// in Helvetica; a character it does not know counts as the widest, so that
// the line never goes past the margin.
func widthOf(text string, s style) float64 {
	total := 0.0
	for _, r := range text {
		if base, ok := unaccented[r]; ok {
			r = base
		}
		switch w, ok := latinWidths[r]; {
		case ok:
			total += w
		case r >= ' ' && r <= '~':
			total += s.widths[r-' ']
		default:
			total += 1000
		}
	}
	return total * s.size / 1000
}

// unaccented are the letters of Spanish and its neighbors that Helvetica
// draws as wide as the letter without the accent.
var unaccented = map[rune]rune{
	'á': 'a', 'à': 'a', 'â': 'a', 'ä': 'a', 'ã': 'a', 'Á': 'A', 'À': 'A', 'Â': 'A', 'Ä': 'A', 'Ã': 'A',
	'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e', 'É': 'E', 'È': 'E', 'Ê': 'E', 'Ë': 'E',
	'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i', 'Í': 'I', 'Ì': 'I', 'Î': 'I', 'Ï': 'I',
	'ó': 'o', 'ò': 'o', 'ô': 'o', 'ö': 'o', 'õ': 'o', 'Ó': 'O', 'Ò': 'O', 'Ô': 'O', 'Ö': 'O', 'Õ': 'O',
	'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u', 'Ú': 'U', 'Ù': 'U', 'Û': 'U', 'Ü': 'U',
	'ñ': 'n', 'Ñ': 'N', 'ç': 'c', 'Ç': 'C',
}

// latinWidths are the signs of Spanish outside ASCII, as wide in Helvetica as
// in Helvetica-Bold.
var latinWidths = map[rune]float64{'¿': 611, '¡': 333, 'º': 365, 'ª': 370, '°': 400, '·': 278}

// helveticaWidths and helveticaBoldWidths are the widths of the characters
// from space to tilde.
var helveticaWidths = [95]float64{
	278, 278, 355, 556, 556, 889, 667, 191, 333, 333, 389, 584, 278, 333, 278, 278, // space to /
	556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 278, 278, 584, 584, 584, 556, // 0 to ?
	1015, 667, 667, 722, 722, 667, 611, 778, 722, 278, 500, 667, 556, 833, 722, 778, // @ to O
	667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 278, 278, 278, 469, 556, // P to _
	333, 556, 556, 500, 556, 556, 278, 556, 556, 222, 222, 500, 222, 833, 556, 556, // ` to o
	556, 556, 333, 500, 278, 556, 500, 722, 500, 500, 500, 334, 260, 334, 584, // p to ~
}

var helveticaBoldWidths = [95]float64{
	278, 333, 474, 556, 556, 889, 722, 238, 333, 333, 389, 584, 278, 333, 278, 278, // space to /
	556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 333, 333, 584, 584, 584, 611, // 0 to ?
	975, 722, 722, 722, 722, 667, 611, 778, 722, 278, 556, 722, 611, 833, 722, 778, // @ to O
	667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 333, 278, 333, 584, 556, // P to _
	333, 556, 611, 556, 611, 556, 333, 611, 611, 278, 278, 556, 278, 889, 611, 611, // ` to o
	611, 611, 389, 556, 333, 611, 556, 778, 556, 556, 500, 389, 280, 389, 584, // p to ~
}
