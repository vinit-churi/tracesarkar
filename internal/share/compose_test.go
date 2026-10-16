package share

import (
	"strings"
	"testing"
	"time"
)

func lodgedFacts() Facts {
	on := time.Date(2026, 6, 11, 8, 14, 0, 0, time.UTC)
	return Facts{
		IssueID:          "8f2a1c",
		Permalink:        "https://tracesarkar.org/issue/8f2a1c",
		Authority:        "BMC",
		Ward:             "R/S",
		Category:         "road_defect",
		Subcategory:      "pothole",
		FirstSeen:        on,
		Sightings:        1,
		Status:           "open",
		Lodged:           true,
		LodgedOn:         &on,
		DeadlineHours:    48,
		DeadlineShortRef: "Bombay HC, PIL 71/2013, para 70(ix)",
		DeadlineCitation: "High Court of Bombay, PIL 71/2013, order of 13 October 2025, paragraph 70(ix).",
		DeadlineState:    "elapsed",
		Handles: []Handle{{
			Platform: "x", Handle: "@mybmc", Kind: Institutional,
			SourceRef: "https://example.test/bmc-handle", RetrievedAt: on,
		}},
	}
}

// An X post that does not fit in the window is not a post. Everything else
// about the kit is worthless if the text cannot be sent.
func TestAnXPostFitsTheWindowInEveryLanguage(t *testing.T) {
	for _, lang := range Languages() {
		p, err := Compose(lodgedFacts(), lang, X)
		if err != nil {
			t.Fatalf("%s: Compose: %v", lang, err)
		}
		if n := len([]rune(p.Text)); n > XLimit {
			t.Errorf("%s: post is %d runes, limit is %d:\n%s", lang, n, XLimit, p.Text)
		}
		if strings.TrimSpace(p.Text) == "" {
			t.Errorf("%s: post is empty", lang)
		}
	}
}

// Hard rule 2. A handle is a fact about a named party, so one that is not in
// the registry cannot reach a post — and the post must carry its source.
func TestNoHandleReachesAPostUnlessTheRegistrySuppliedIt(t *testing.T) {
	f := lodgedFacts()
	f.Handles = nil
	p, err := Compose(f, EN, X)
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	if strings.Contains(p.Text, "@") {
		t.Errorf("a post with no registry handle still tagged somebody:\n%s", p.Text)
	}

	p, err = Compose(lodgedFacts(), EN, X)
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	if !strings.Contains(p.Text, "@mybmc") {
		t.Errorf("the registry handle did not reach the post:\n%s", p.Text)
	}
	if len(p.Sources) == 0 {
		t.Error("a post naming a party carries no source reference")
	}
}

// An escalation handle on a first report trains everybody to ignore the
// platform (share kit §4), so it is unlocked by an elapsed deadline only.
func TestAnEscalationHandleWaitsForAnElapsedDeadline(t *testing.T) {
	esc := Handle{Platform: "x", Handle: "@CMOMaharashtra", Kind: Escalation,
		SourceRef: "https://example.test/cmo", RetrievedAt: time.Now()}

	running := lodgedFacts()
	running.DeadlineState = "running"
	running.Handles = append(running.Handles, esc)
	p, err := Compose(running, EN, X)
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	if strings.Contains(p.Text, "@CMOMaharashtra") {
		t.Errorf("escalation handle tagged while the deadline was still running:\n%s", p.Text)
	}

	elapsed := lodgedFacts()
	elapsed.Handles = append(elapsed.Handles, esc)
	p, err = Compose(elapsed, EN, WhatsApp)
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	if !strings.Contains(p.Text, "@CMOMaharashtra") {
		t.Errorf("escalation handle withheld after the deadline elapsed:\n%s", p.Text)
	}
}

// A deadline shown without the document that creates it is an assertion, and
// this platform does not make those. Dropping the sentence is correct; showing
// the number alone is not.
func TestADeadlineWithNoCitationIsNotStatedAtAll(t *testing.T) {
	f := lodgedFacts()
	f.DeadlineShortRef, f.DeadlineCitation = "", ""
	p, err := Compose(f, EN, WhatsApp)
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	if strings.Contains(p.Text, "48") {
		t.Errorf("an uncited deadline reached the post:\n%s", p.Text)
	}
}

// Hard rule 7 of the share kit: photographing something here notifies nobody.
// A post about an unlodged issue must not imply the platform filed it.
func TestAnUnlodgedIssueDoesNotClaimAnybodyWasTold(t *testing.T) {
	f := lodgedFacts()
	f.Lodged, f.LodgedOn, f.DeadlineState = false, nil, "not_started"
	p, err := Compose(f, EN, WhatsApp)
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	low := strings.ToLower(p.Text)
	for _, claim := range []string{"lodged", "we have reported", "filed", "complaint has been"} {
		if strings.Contains(low, claim) {
			t.Errorf("an unlodged issue claimed %q:\n%s", claim, p.Text)
		}
	}
}

// The legal citation is the one thing that must survive every language
// unchanged (vernacular doc §1 rule 3, and the no-machine-translated-legal-
// text rule).
func TestTheCitationIsNeverTranslated(t *testing.T) {
	for _, lang := range Languages() {
		p, err := Compose(lodgedFacts(), lang, WhatsApp)
		if err != nil {
			t.Fatalf("%s: Compose: %v", lang, err)
		}
		if !strings.Contains(p.Text, "PIL 71/2013") {
			t.Errorf("%s: the citation did not survive into the post:\n%s", lang, p.Text)
		}
	}
}

// Generated per language from the fact set, never translated from English.
// Two languages producing the same bytes means one of them is a copy.
func TestEachLanguageIsGeneratedNotCopied(t *testing.T) {
	seen := map[string]Lang{}
	for _, lang := range Languages() {
		p, err := Compose(lodgedFacts(), lang, WhatsApp)
		if err != nil {
			t.Fatalf("%s: Compose: %v", lang, err)
		}
		if prev, dup := seen[p.Text]; dup {
			t.Errorf("%s and %s produced identical text", prev, lang)
		}
		seen[p.Text] = lang
	}
}

// The glossary is explicit: a prabhag is an electoral division and is NOT an
// administrative ward. Marathi copy that calls a ward a prabhag is wrong in
// the one way a BMC desk would notice.
func TestMarathiDoesNotCallAnAdministrativeWardAPrabhag(t *testing.T) {
	p, err := Compose(lodgedFacts(), MR, WhatsApp)
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	if strings.Contains(p.Text, "प्रभाग") {
		t.Errorf("Marathi post called an administrative ward a prabhag:\n%s", p.Text)
	}
}

// An unreviewed translation may be offered as a draft, but never silently: a
// caller has to be able to tell it apart from reviewed copy.
//
// This asserts the mechanism rather than which languages happen to be
// reviewed today. Marathi and Hindi were drafts until 16 October 2026 and are
// not now ([D175](../../docs/00-overview/05-decision-log.md)); a test that
// hard-coded that state would have to be edited every time a language is
// signed off, which is exactly when nobody wants to be editing tests.
func TestEveryPostCarriesTheReviewStateOfTheFileBehindIt(t *testing.T) {
	states := Review()
	for _, lang := range Languages() {
		p, err := Compose(lodgedFacts(), lang, WhatsApp)
		if err != nil {
			t.Fatalf("Compose(%s): %v", lang, err)
		}
		switch p.Review {
		case Reviewed, Draft:
		default:
			t.Errorf("%s carries review state %q, which is neither known value",
				lang, p.Review)
		}
		if p.Review != states[lang] {
			t.Errorf("%s composed as %q but Review() says %q",
				lang, p.Review, states[lang])
		}
	}
	// English is the language the other two are checked against; it cannot be
	// a draft.
	if states[EN] != Reviewed {
		t.Errorf("English is %q, want %q", states[EN], Reviewed)
	}
}

// The cautious default, which is the half of this that cannot be read off the
// files: anything that does not say "reviewed" is a draft, so a new language
// cannot ship as reviewed by omission or by a typo in that field.
func TestAnythingButReviewedIsADraft(t *testing.T) {
	for _, word := range []string{"", "pending", "Reviewed", "reviewed ", "in progress", "yes"} {
		if got := (&vocab{Review: word}).state(); got != Draft {
			t.Errorf("review %q yielded %q, want %q", word, got, Draft)
		}
	}
	if got := (&vocab{Review: "reviewed"}).state(); got != Reviewed {
		t.Errorf("review \"reviewed\" yielded %q", got)
	}
}

// G14, restated for generated text: these are different claims by different
// parties and must never read the same.
func TestAnAuthorityClaimAndACitizenConfirmationNeverReadTheSame(t *testing.T) {
	claimed, citizen := lodgedFacts(), lodgedFacts()
	claimed.Status, citizen.Status = "claimed_resolved", "citizen_confirmed"
	a, err := Compose(claimed, EN, WhatsApp)
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	b, err := Compose(citizen, EN, WhatsApp)
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	if a.Text == b.Text {
		t.Error("claimed_resolved and citizen_confirmed produced the same post")
	}
}

// Every channel must produce something for every language, or the kit has
// holes the UI has to special-case.
func TestEveryChannelAndLanguagePairProducesAPost(t *testing.T) {
	for _, lang := range Languages() {
		for _, ch := range Channels() {
			p, err := Compose(lodgedFacts(), lang, ch)
			if err != nil {
				t.Errorf("%s/%s: %v", lang, ch, err)
				continue
			}
			if !strings.Contains(p.Text, "tracesarkar.org/issue/8f2a1c") {
				t.Errorf("%s/%s: the permalink is missing — a forwarded post "+
					"without provenance is a rumour:\n%s", lang, ch, p.Text)
			}
		}
	}
}

func TestComposeRefusesAnUnknownLanguageOrChannel(t *testing.T) {
	if _, err := Compose(lodgedFacts(), Lang("fr"), X); err == nil {
		t.Error("Compose accepted a language with no vocabulary")
	}
	if _, err := Compose(lodgedFacts(), EN, Channel("telegram")); err == nil {
		t.Error("Compose accepted an unknown channel")
	}
}
