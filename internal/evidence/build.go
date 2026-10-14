package evidence

import (
	"bytes"
	"fmt"
	"strings"
)

// Build compiles the record into a PDF.
//
// Photographs come first after the cover, because the photograph is the
// evidence and everything else is the apparatus around it. They are the
// redacted copies: this document exists to be handed to somebody else, which
// is exactly the case redaction is for.
func Build(r Record, photos []Photo) ([]byte, error) {
	p := &pdf{}

	// The two standard faces, which every reader already has, so nothing is
	// embedded and the file stays small enough to email.
	regular := p.addf("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>")
	boldFont := p.addf("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>")

	images := make([]int, 0, len(photos))
	for i, ph := range photos {
		if len(ph.JPEG) == 0 || ph.Width <= 0 || ph.Height <= 0 {
			return nil, fmt.Errorf("evidence: photograph %d has no image", i+1)
		}
		images = append(images, p.stream(fmt.Sprintf(
			"/Type /XObject /Subtype /Image /Width %d /Height %d "+
				"/ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode",
			ph.Width, ph.Height), ph.JPEG))
	}

	lay := &layout{stamp: Stamp(r), issueID: r.IssueID}
	lay.start()
	lay.title(titleOf(r))
	lay.sub("Compiled record · issue " + r.IssueID)
	lay.sub("Stamp " + lay.stamp)
	lay.rule()

	for i, ph := range photos {
		lay.photo(i, ph)
	}
	for _, b := range blocks(r) {
		lay.block(b)
	}
	lay.finish()

	kids := make([]string, 0, len(lay.pages))
	pagesObj := len(p.objects) + 1 + 2*len(lay.pages) // reserved below
	for _, page := range lay.pages {
		content := p.stream("", page.content.Bytes())
		res := "/Font << /F1 " + ref(regular) + " /F2 " + ref(boldFont) + " >>"
		if len(page.images) > 0 {
			var xo []string
			for _, idx := range page.images {
				xo = append(xo, fmt.Sprintf("/Im%d %s", idx, ref(images[idx])))
			}
			res += " /XObject << " + strings.Join(xo, " ") + " >>"
		}
		obj := p.addf("<< /Type /Page\n /Parent %s /MediaBox [0 0 %.2f %.2f] "+
			"/Resources << %s >> /Contents %s >>",
			ref(pagesObj), pageWidth, pageHeight, res, ref(content))
		kids = append(kids, ref(obj))
	}

	got := p.addf("<< /Type /Pages /Kids [%s] /Count %d >>",
		strings.Join(kids, " "), len(kids))
	if got != pagesObj {
		// The page tree's object number is referenced by every page before it
		// exists, so it is reserved by arithmetic. If that arithmetic is ever
		// wrong the file is subtly broken in a way no reader reports clearly,
		// so it is checked rather than trusted.
		return nil, fmt.Errorf("evidence: page tree landed at object %d, expected %d", got, pagesObj)
	}
	root := p.addf("<< /Type /Catalog /Pages %s >>", ref(got))
	return p.render(root), nil
}

func ref(n int) string { return fmt.Sprintf("%d 0 R", n) }

func titleOf(r Record) string {
	thing := strings.TrimSpace(r.Subcategory)
	if thing == "" {
		thing = strings.ReplaceAll(r.Category, "_", " ")
	}
	t := strings.ToUpper(thing[:1]) + thing[1:]
	if r.Ward != "" {
		t += " · " + r.Ward + " ward"
	}
	return t
}

type page struct {
	content bytes.Buffer
	images  []int
}

// layout walks down the page and starts a new one when it runs out of room.
type layout struct {
	pages   []*page
	cur     *page
	y       float64
	stamp   string
	issueID string
}

func (l *layout) start() {
	l.cur = &page{}
	l.pages = append(l.pages, l.cur)
	l.y = pageHeight - marginTop
}

// room reports whether h points remain above the footer, starting a new page
// if not.
func (l *layout) room(h float64) {
	if l.y-h >= marginFoot {
		return
	}
	l.footer()
	l.start()
}

func (l *layout) text(s string, f face, size float64) {
	name := "F1"
	if f == boldFace {
		name = "F2"
	}
	fmt.Fprintf(&l.cur.content, "BT /%s %.2f Tf 1 0 0 1 %.2f %.2f Tm %s Tj ET\n",
		name, size, marginLeft, l.y, pdfString(s))
}

func (l *layout) title(s string) {
	l.room(titleSize * 1.4)
	l.y -= titleSize
	l.text(s, boldFace, titleSize)
	l.y -= 6
}

func (l *layout) sub(s string) {
	l.room(leadSize)
	l.y -= smallSize + 3
	l.text(s, bodyFace, smallSize)
}

func (l *layout) rule() {
	l.room(14)
	l.y -= 10
	fmt.Fprintf(&l.cur.content, "0.78 0.76 0.73 RG 0.6 w %.2f %.2f m %.2f %.2f l S\n",
		marginLeft, l.y, pageWidth-marginRight, l.y)
	l.y -= 6
}

func (l *layout) heading(s string) {
	l.room(headSize * 3)
	l.y -= headSize * 1.9
	l.text(strings.ToUpper(s), boldFace, headSize)
	l.y -= 4
}

func (l *layout) paragraph(s string, f face) {
	for _, ln := range wrap(s, f, bodySize, usableWidth) {
		l.room(leadSize)
		l.y -= leadSize
		l.text(ln, f, bodySize)
	}
}

func (l *layout) block(b block) {
	if b.heading != "" {
		l.heading(b.heading)
	}
	for _, ln := range b.lines {
		if ln.gap {
			l.y -= 6
			continue
		}
		l.paragraph(ln.text, ln.style)
	}
}

// photo places one capture at the page width it fits, with what is known about
// it underneath — including which detector obscured what, because a redacted
// photograph presented without that record looks like a damaged one.
func (l *layout) photo(idx int, ph Photo) {
	w := pageWidth - marginLeft - marginRight
	h := w * float64(ph.Height) / float64(ph.Width)
	if maxH := pageHeight - marginTop - marginFoot - 90; h > maxH {
		h = maxH
		w = h * float64(ph.Width) / float64(ph.Height)
	}
	l.room(h + 44)
	l.y -= h
	fmt.Fprintf(&l.cur.content, "q %.2f 0 0 %.2f %.2f %.2f cm /Im%d Do Q\n",
		w, h, marginLeft, l.y, idx)
	l.cur.images = append(l.cur.images, idx)

	if !ph.TakenAt.IsZero() {
		l.paragraph("Photographed "+moment(ph.TakenAt)+".", bodyFace)
	}
	if ph.AccuracyM > 0 {
		l.paragraph(fmt.Sprintf("Position accurate to about %.0f m at the moment of capture.",
			ph.AccuracyM), bodyFace)
	}
	if strings.TrimSpace(ph.Redaction) != "" {
		l.paragraph("Faces and number plates were obscured before this copy existed: "+
			ph.Redaction+". The archival original is not reproduced here.", bodyFace)
	}
}

func (l *layout) footer() {
	y := marginFoot - 18
	fmt.Fprintf(&l.cur.content,
		"BT /F1 %.2f Tf 1 0 0 1 %.2f %.2f Tm %s Tj ET\n",
		smallSize-0.6, marginLeft, y,
		pdfString("TraceSarkar · compiled record · issue "+l.issueID+" · stamp "+l.stamp))
}

func (l *layout) finish() { l.footer() }
