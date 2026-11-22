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
	bold    bool
}

var (
	titleStyle   = style{font: "F2", size: 16, leading: 22, bold: true}
	headingStyle = style{font: "F2", size: 12, leading: 18, bold: true}
	labelStyle   = style{font: "F2", size: 10, leading: 14, bold: true}
	textStyle    = style{font: "F1", size: 10, leading: 14}
	noteStyle    = style{font: "F1", size: 8, leading: 11, gray: 0.4}
)

// Document lays out its lines top to bottom and starts a new page when one
// does not fit. Every page ends with the footer and its number.
type Document struct {
	title  string
	footer string
	pages  []*bytes.Buffer
	y      float64
}

func New(title, footer string) *Document {
	d := &Document{title: title, footer: footer}
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

// widthOf measures text with the widths of Helvetica, in thousandths of the
// font size. Helvetica-Bold is a little wider, so it counts a tenth more.
func widthOf(text string, s style) float64 {
	total := 0.0
	for _, b := range []byte(encode(text)) {
		w := 778.0 // the widest Latin letters, for those not in the table
		if b >= ' ' && b <= '~' {
			w = helveticaWidths[b-' ']
		}
		total += w
	}
	if s.bold {
		total *= 1.1
	}
	return total * s.size / 1000
}

// helveticaWidths are the widths of the characters from space to tilde.
var helveticaWidths = [...]float64{
	278, 278, 355, 556, 556, 889, 667, 191, 333, 333, 389, 584, 278, 333, 278, 278, // space to /
	556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 278, 278, 584, 584, 584, 556, // 0 to ?
	1015, 667, 667, 722, 722, 667, 611, 778, 722, 278, 500, 667, 556, 833, 722, 778, // @ to O
	667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 278, 278, 278, 469, 556, // P to _
	333, 556, 556, 500, 556, 556, 278, 556, 556, 222, 222, 500, 222, 833, 556, 556, // ` to o
	556, 556, 333, 500, 278, 556, 500, 722, 500, 500, 500, 334, 260, 334, 584, // p to ~
}
