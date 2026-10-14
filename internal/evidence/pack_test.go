package evidence

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"strings"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/share"
)

func aPhoto(t *testing.T) Photo {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 640, 480))
	for y := 0; y < 480; y++ {
		for x := 0; x < 640; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: 90, B: 120, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return Photo{
		JPEG:      buf.Bytes(),
		Width:     640,
		Height:    480,
		TakenAt:   time.Date(2026, 10, 11, 9, 14, 0, 0, time.UTC),
		AccuracyM: 6.2,
		Redaction: "pigo-1.4.6/facefinder+edge-density-2, 1 region obscured",
	}
}

func aRecord() Record {
	lodged := time.Date(2026, 10, 12, 11, 0, 0, 0, time.UTC)
	return Record{
		IssueID:     "174642cf-f19c-4c5b-b047-31bbc12b865e",
		Category:    "road_defect",
		Subcategory: "pothole",
		Status:      "open",
		Sightings:   2,
		Authority:   "BMC",
		Ward:        "R/C",
		WardBasis:   "The point lies inside the R/C boundary with 41 m to the nearest edge",
		Lat:         19.2309441,
		Lon:         72.8567123,
		AccuracyM:   6.2,
		FirstSeen:   time.Date(2026, 10, 11, 9, 14, 0, 0, time.UTC),
		LastSeen:    time.Date(2026, 10, 12, 8, 2, 0, 0, time.UTC),
		LodgedOn:    &lodged,
		Reference:   "BMC/RC/2026/44871",
		Desk: &Desk{
			Department:  "Chief Engineer (Roads & Traffic) — Western Suburbs",
			Officer:     "Assistant Engineer (Roads) R/Central",
			Address:     "Office of Dy. Ch. Eng. Roads (W.S.), S. V. Road, Goregaon (West)",
			SourceRef:   "https://portal.mcgm.gov.in/example.pdf",
			RetrievedAt: time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC),
		},
		DeadlineHours:    48,
		DeadlineCitation: `High Court of Bombay, PIL 71/2013, order of 13 October 2025, paragraph 70(ix).`,
		Events: []Event{
			{Seq: 1, Kind: "reported", At: time.Date(2026, 10, 11, 9, 14, 0, 0, time.UTC),
				Detail: "No open road defect issue lies within 8 m", Hash: "9f2ca41b"},
			{Seq: 2, Kind: "joined", At: time.Date(2026, 10, 12, 8, 2, 0, 0, time.UTC),
				Detail: "A second sighting 4 m away", Hash: "b6132f09"},
		},
	}
}

func text(pdf []byte) string { return string(pdf) }

// A PDF that no reader opens is not a document.
func TestThePackIsAWellFormedPDF(t *testing.T) {
	out, err := Build(aRecord(), nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF-1.4")) {
		t.Errorf("no PDF header: %q", out[:min(16, len(out))])
	}
	if !bytes.Contains(out, []byte("\nxref\n")) {
		t.Error("no cross-reference table")
	}
	if !bytes.Contains(out, []byte("startxref")) {
		t.Error("no startxref")
	}
	if !bytes.HasSuffix(bytes.TrimRight(out, "\n"), []byte("%%EOF")) {
		t.Error("no EOF marker")
	}
}

// Two people generating the pack for the same issue must get the same file, or
// the hash printed on it proves nothing and the document cannot be cited.
func TestThePackIsDeterministic(t *testing.T) {
	a, err := Build(aRecord(), []Photo{aPhoto(t)})
	if err != nil {
		t.Fatal(err)
	}
	b, err := Build(aRecord(), []Photo{aPhoto(t)})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Error("two runs produced different bytes")
	}
	// A wall-clock timestamp is the usual way this breaks.
	if bytes.Contains(a, []byte("/CreationDate")) {
		t.Error("the pack carries a creation date, so it cannot be deterministic")
	}
}

// The stamp is what lets two people agree they are holding the same record.
func TestThePackCarriesAStampThatChangesWithTheFacts(t *testing.T) {
	a := aRecord()
	first := Stamp(a)
	if len(first) < 8 {
		t.Fatalf("stamp is too short to mean anything: %q", first)
	}
	if !strings.Contains(text(mustBuild(t, a, nil)), first) {
		t.Error("the stamp is not printed in the document")
	}

	b := aRecord()
	b.Status = "claimed_resolved"
	if Stamp(b) == first {
		t.Error("changing the status did not change the stamp")
	}

	c := aRecord()
	c.Events[1].Detail = "A second sighting 5 m away"
	if Stamp(c) == first {
		t.Error("changing the history did not change the stamp")
	}
}

// Hard rule 3, in the document most likely to be attached to a filing.
func TestThePackAssertsNoWrongdoing(t *testing.T) {
	for _, st := range []string{"open", "claimed_resolved", "citizen_confirmed", "reopened"} {
		r := aRecord()
		r.Status = st
		for _, line := range Lines(r) {
			if vs := share.Check(line); len(vs) != 0 {
				t.Errorf("%s: %q fails the tone check: %+v", st, line, vs)
			}
		}
	}
}

// It is a compiled record, not advice, and it has to say so.
func TestThePackSaysWhatItIsNot(t *testing.T) {
	body := strings.Join(Lines(aRecord()), "\n")
	for _, want := range []string{"not legal advice", "compiled"} {
		if !strings.Contains(strings.ToLower(body), want) {
			t.Errorf("the pack does not say it is %q", want)
		}
	}
}

// Hard rule 2. The desk is a named party and the deadline is a legal claim;
// neither appears without the document behind it.
func TestEveryNamedPartyFactCarriesItsSource(t *testing.T) {
	body := strings.Join(Lines(aRecord()), "\n")
	if !strings.Contains(body, "portal.mcgm.gov.in/example.pdf") {
		t.Error("the desk is named with no source reference")
	}
	if !strings.Contains(body, "29 September 2026") {
		t.Error("the desk's source has no retrieval date")
	}
	if !strings.Contains(body, "PIL 71/2013") {
		t.Error("the deadline is stated with no citation")
	}

	// A desk with no source is not a desk the pack will name.
	r := aRecord()
	r.Desk.SourceRef = ""
	body = strings.Join(Lines(r), "\n")
	if strings.Contains(body, "Chief Engineer (Roads & Traffic)") {
		t.Error("an unsourced desk was named anyway")
	}
}

// A contract is the platform's central claim and the expensive error. Nothing
// is published today, so the pack must say so rather than leave a gap.
func TestAnUnattributedIssueSaysSoRatherThanLeavingAGap(t *testing.T) {
	body := strings.Join(Lines(aRecord()), "\n")
	if !strings.Contains(strings.ToLower(body), "no contract") {
		t.Errorf("the pack is silent about contract attribution:\n%s", body)
	}

	r := aRecord()
	r.Contract = &Contract{
		WorkCode: "WS/2023/ROAD/117", Contractor: "Example Infra Pvt Ltd",
		Confidence: "high", DistanceM: 7.4,
		Basis:       "The point lies 7.4 m from the centreline of this work",
		SourceRef:   "https://roads.mcgm.gov.in/api/works/117",
		RetrievedAt: time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC),
	}
	body = strings.Join(Lines(r), "\n")
	if !strings.Contains(body, "WS/2023/ROAD/117") {
		t.Error("a published contract did not reach the pack")
	}
	if !strings.Contains(body, "24 August 2026") {
		t.Error("the contract source has no retrieval date")
	}

	// Without a source it is not printed, whatever the confidence says.
	r.Contract.SourceRef = ""
	body = strings.Join(Lines(r), "\n")
	if strings.Contains(body, "Example Infra Pvt Ltd") {
		t.Error("a contractor was named with no source")
	}
}

// This pack is for the person who reported it, to attach to a filing. The
// exact position is what a complaint needs, and it is labelled as exact so it
// is never mistaken for the coarsened one the public page shows.
func TestThePackCarriesTheExactPositionAndSaysThatIsWhatItIs(t *testing.T) {
	body := strings.Join(Lines(aRecord()), "\n")
	if !strings.Contains(body, "19.2309441") || !strings.Contains(body, "72.8567123") {
		t.Error("the filing document does not carry the position it needs")
	}
	if !strings.Contains(strings.ToLower(body), "not the coarsened") {
		t.Error("the pack does not distinguish this from the published position")
	}
}

func TestThePhotographIsEmbeddedWithItsRedactionRecord(t *testing.T) {
	out := mustBuild(t, aRecord(), []Photo{aPhoto(t)})
	if !bytes.Contains(out, []byte("/DCTDecode")) {
		t.Error("the photograph was not embedded")
	}
	if !bytes.Contains(out, []byte("/XObject")) {
		t.Error("no image object in the page resources")
	}
	body := strings.Join(Lines(aRecord()), "\n")
	_ = body
	if !bytes.Contains(out, []byte("edge-density-2")) {
		t.Error("the redaction record does not travel with the photograph")
	}
}

// Text longer than a line has to wrap rather than run off the page.
func TestLongTextWrapsInsideThePage(t *testing.T) {
	r := aRecord()
	r.WardBasis = strings.Repeat("The point lies inside the boundary. ", 40)
	out := mustBuild(t, r, nil)
	if len(out) == 0 {
		t.Fatal("no document")
	}
	for _, line := range wrap(r.WardBasis, bodyFace, bodySize, usableWidth) {
		if w := measure(line, bodyFace, bodySize); w > usableWidth {
			t.Errorf("line is %.1fpt wide, usable width is %.1fpt: %q", w, usableWidth, line)
		}
	}
}

// Enough content has to run onto a second page rather than off the first.
func TestThePackRunsOntoFurtherPagesRatherThanOffTheFirst(t *testing.T) {
	r := aRecord()
	for i := 0; i < 60; i++ {
		r.Events = append(r.Events, Event{
			Seq: i + 3, Kind: "re-checked",
			At:     time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC).Add(time.Duration(i) * time.Hour),
			Detail: "Photographed again and still present", Hash: "aaaa0000",
		})
	}
	out := mustBuild(t, r, nil)
	if n := bytes.Count(out, []byte("/Type /Page\n")); n < 2 {
		t.Errorf("the pack is %d page(s); sixty events cannot fit on one", n)
	}
}

func mustBuild(t *testing.T, r Record, ps []Photo) []byte {
	t.Helper()
	out, err := Build(r, ps)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// A source reference that runs off the page is not a source reference. Hard
// rule 2 exists so a reader can go and check; a URL cut at the margin cannot
// be checked, and this is the document most likely to be printed.
func TestALongSourceUrlIsBrokenRatherThanRunOffThePage(t *testing.T) {
	long := "https://portal.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/" +
		"Roads%20and%20Traffic/RTI%20Manuals/Westernsuburbs/Westernsuburbs_RTI_E01.pdf"
	for _, line := range wrap("Source: "+long+", retrieved 29 September 2026.",
		bodyFace, bodySize, usableWidth) {
		if w := measure(line, bodyFace, bodySize); w > usableWidth {
			t.Errorf("line is %.1fpt wide, usable width is %.1fpt: %q", w, usableWidth, line)
		}
	}

	// Broken, not altered: putting the pieces back must give the URL again.
	var rebuilt string
	for _, line := range wrap(long, bodyFace, bodySize, usableWidth) {
		rebuilt += strings.TrimSpace(line)
	}
	if rebuilt != long {
		t.Errorf("the URL did not survive being broken:\n got %s\nwant %s", rebuilt, long)
	}
}

// Sentences that already end in a full stop must not get a second one.
func TestNoSentenceEndsInTwoFullStops(t *testing.T) {
	r := aRecord()
	r.WardBasis = "The point lies inside BMC R/C, 873 m from the nearest ward line."
	r.Contract = &Contract{
		WorkCode: "W-447", Contractor: "Example Infra",
		Basis:       "The point lies 6.5 m from the centreline.",
		SourceRef:   "https://example.test/w447",
		RetrievedAt: time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC),
	}
	for _, line := range Lines(r) {
		if strings.Contains(line, "..") && !strings.Contains(line, "...") {
			t.Errorf("double full stop: %q", line)
		}
	}
}
