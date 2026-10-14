// Package evidence compiles one issue into a document a person can take to a
// meeting or attach to a filing.
//
// Two properties shape everything here.
//
// It is **deterministic**: two people generating the pack for the same issue
// get byte-identical files, so the stamp printed on it means something and the
// document can be cited. That rules out a creation timestamp, a random
// document id, and map iteration order — the usual three ways a PDF writer
// becomes irreproducible without anybody noticing.
//
// And it is **compiled, not argued**: every fact in it is one the platform
// already holds, each named party carries the document behind it, and the pack
// states plainly that it is a record rather than advice.
//
// The writer is written here rather than taken from a library because the
// determinism above is the whole point and most PDF libraries stamp a creation
// date and a generated /ID by default. The subset needed is small: the
// standard Helvetica faces, which need no embedding, and JPEG images, which a
// PDF can carry as-is.
package evidence

import (
	"bytes"
	"fmt"
	"strings"
)

// Page geometry, in points. A4, because a filing in Mumbai is on A4.
const (
	pageWidth   = 595.28
	pageHeight  = 841.89
	marginLeft  = 56.0
	marginRight = 56.0
	marginTop   = 56.0
	marginFoot  = 48.0

	// usableWidth is what a line of body text may occupy. The 0.98 is slack
	// against the width table below: it is the published Helvetica metric set,
	// and a renderer that rounds differently should still not overrun.
	usableWidth = (pageWidth - marginLeft - marginRight) * 0.98
)

type face int

const (
	bodyFace face = iota
	boldFace
)

const (
	bodySize  = 9.6
	leadSize  = 11.5
	titleSize = 20.0
	headSize  = 11.0
	smallSize = 8.2
)

// pdf accumulates objects and serialises them.
//
// Object numbers are handed out in the order objects are added, and nothing
// iterates a map while writing, so the same input always produces the same
// bytes.
type pdf struct {
	objects [][]byte
}

func (p *pdf) add(body []byte) int {
	p.objects = append(p.objects, body)
	return len(p.objects) // 1-based; object 0 is the free head
}

func (p *pdf) addf(format string, args ...any) int {
	return p.add([]byte(fmt.Sprintf(format, args...)))
}

// stream writes a stream object with its dictionary.
func (p *pdf) stream(dict string, data []byte) int {
	var b bytes.Buffer
	fmt.Fprintf(&b, "<< %s /Length %d >>\nstream\n", dict, len(data))
	b.Write(data)
	b.WriteString("\nendstream")
	return p.add(b.Bytes())
}

// render serialises the document. root is the catalogue's object number.
func (p *pdf) render(root int) []byte {
	var b bytes.Buffer
	// 1.4 is old enough that every reader and every government portal opens it.
	b.WriteString("%PDF-1.4\n")
	// A binary comment line, so anything moving the file treats it as binary.
	b.Write([]byte{'%', 0xE2, 0xE3, 0xCF, 0xD3, '\n'})

	offsets := make([]int, len(p.objects)+1)
	for i, body := range p.objects {
		offsets[i+1] = b.Len()
		fmt.Fprintf(&b, "%d 0 obj\n", i+1)
		b.Write(body)
		b.WriteString("\nendobj\n")
	}

	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n", len(p.objects)+1)
	b.WriteString("0000000000 65535 f \n")
	for i := 1; i <= len(p.objects); i++ {
		fmt.Fprintf(&b, "%010d 00000 n \n", offsets[i])
	}
	// No /ID and no /CreationDate: both are conventionally random or
	// time-based, and either would make the file different on every run.
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root %d 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		len(p.objects)+1, root, xref)
	return b.Bytes()
}

// pdfString escapes a string for a PDF literal.
//
// WinAnsi is a single-byte encoding, so anything outside it is transliterated
// rather than emitted raw — a stray byte would render as a wrong glyph, which
// in a document meant for a filing is worse than a plain substitute.
func pdfString(s string) string {
	var b strings.Builder
	b.WriteByte('(')
	for _, r := range s {
		switch r {
		case '(', ')', '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '\n', '\r', '\t':
			b.WriteByte(' ')
		default:
			if r < 32 {
				b.WriteByte(' ')
			} else if r < 127 {
				b.WriteRune(r)
			} else if sub, ok := transliterate[r]; ok {
				b.WriteString(sub)
			} else {
				b.WriteByte('?')
			}
		}
	}
	b.WriteByte(')')
	return b.String()
}

// transliterate covers the non-ASCII this document actually contains: the
// typographic marks the platform's own copy uses, and the rupee sign.
var transliterate = map[rune]string{
	'·': "-",   // middle dot
	'—': " - ", // em dash
	'–': "-",   // en dash
	'‘': "'", '’': "'",
	'“': "\"", '”': "\"",
	'…': "...",
	'₹': "Rs ", // rupee
	' ': " ",
	'→': "->",
	'×': "x",
}

// measure returns the width of a string at a size, in points.
func measure(s string, f face, size float64) float64 {
	w := widths(f)
	var total float64
	for _, r := range s {
		idx := int(r)
		if idx < 32 || idx > 126 {
			// Everything outside ASCII is transliterated before it is drawn;
			// a question mark is the widest thing it can become.
			idx = '?'
		}
		total += float64(w[idx-32])
	}
	return total * size / 1000
}

// wrap breaks text into lines that fit a width, on word boundaries.
//
// A word longer than a whole line is broken too, because in this document the
// over-long words are source URLs, and a URL that runs off the right margin
// cannot be checked — which is the entire reason hard rule 2 requires one. It
// is broken and never hyphenated: a hyphen inserted into a URL becomes part of
// it when somebody types it back in.
func wrap(s string, f face, size, width float64) []string {
	var lines []string
	line := ""
	flush := func() {
		if line != "" {
			lines = append(lines, line)
			line = ""
		}
	}
	for _, word := range strings.Fields(s) {
		candidate := word
		if line != "" {
			candidate = line + " " + word
		}
		if measure(candidate, f, size) <= width {
			line = candidate
			continue
		}
		flush()
		if measure(word, f, size) <= width {
			line = word
			continue
		}
		for _, piece := range breakLong(word, f, size, width) {
			flush()
			line = piece
		}
	}
	flush()
	return lines
}

// breakLong splits a word too wide for any line.
//
// It prefers to break after a path separator, so a URL comes apart where a
// reader already expects a seam and reads back as the same address. Nothing is
// inserted: joining the pieces returns the original word exactly.
func breakLong(word string, f face, size, width float64) []string {
	var out []string
	cur := ""
	lastSeam := -1
	for _, r := range word {
		next := cur + string(r)
		if measure(next, f, size) > width && cur != "" {
			cut := len(cur)
			if lastSeam > 0 {
				cut = lastSeam
			}
			out = append(out, cur[:cut])
			cur = cur[cut:] + string(r)
			lastSeam = -1
			continue
		}
		cur = next
		if r == '/' || r == '&' || r == '?' || r == '=' || r == '_' {
			lastSeam = len(cur)
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
