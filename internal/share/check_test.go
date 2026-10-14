package share

import (
	"strings"
	"testing"
)

// The prohibited-pattern list the share kit spec §3 requires, run after
// generation rather than trusted to the template. The template is ours today;
// the check is what holds when somebody edits one.
func TestCheckRefusesTextThatAssertsWrongdoing(t *testing.T) {
	cases := []struct {
		name string
		text string
		want string
	}{
		{"corruption", "Pothole in R/S. This is corruption by BMC.", "wrongdoing"},
		{"scam", "A ₹4 cr scam on this road.", "wrongdoing"},
		{"embezzled", "The contractor embezzled the road budget.", "wrongdoing"},
		{"bribe", "Nothing moves here without a bribe.", "wrongdoing"},
		{"negligence as a charge", "BMC is guilty of negligence.", "wrongdoing"},
		{"insult", "BMC engineers are useless idiots.", "insult"},
		{"sarcasm", "Great job as usual, what a joke.", "insult"},
		{"party reference", "This is what the BJP government delivers.", "political"},
		{"another party", "Shiv Sena ran this ward into the ground.", "political"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vs := Check(c.text)
			if len(vs) == 0 {
				t.Fatalf("Check passed text it should refuse: %q", c.text)
			}
			var kinds []string
			for _, v := range vs {
				kinds = append(kinds, v.Kind)
			}
			if !contains(kinds, c.want) {
				t.Errorf("Check found %v, want a %q violation", kinds, c.want)
			}
		})
	}
}

// The check must not fire on the platform's own vocabulary. A validator that
// refuses every real post gets switched off.
func TestCheckPassesTheFactsThePlatformActuallyStates(t *testing.T) {
	ok := []string{
		"Pothole in R/S ward. @mybmc — this is your ward. 48 hours to attend " +
			"(Bombay HC, PIL 71/2013, para 70(ix)) — passed 3 days ago.",
		"The authority has recorded this as attended. No resident has " +
			"photographed it fixed.",
		"Under warranty from contract WS/2023/ROAD/117 until 28 Nov 2028.",
		"Seen 7 times since 11 Jun 2026.",
		"Nobody has recorded telling the authority about this yet, so no " +
			"deadline is running.",
	}
	for _, s := range ok {
		if vs := Check(s); len(vs) != 0 {
			t.Errorf("Check refused legitimate text %q: %+v", s, vs)
		}
	}
}

// Every post the package produces has to pass its own check, in every
// language and channel. If a template cannot, that is a bug in the template.
func TestEveryGeneratedPostPassesItsOwnCheck(t *testing.T) {
	for _, st := range []string{"open", "claimed_resolved", "citizen_confirmed", "reopened"} {
		for _, lodged := range []bool{true, false} {
			for _, lang := range Languages() {
				for _, ch := range Channels() {
					f := lodgedFacts()
					f.Status = st
					if !lodged {
						f.Lodged, f.LodgedOn, f.DeadlineState = false, nil, "not_started"
					}
					p, err := Compose(f, lang, ch)
					if err != nil {
						t.Fatalf("%s/%s: %v", lang, ch, err)
					}
					if vs := Check(p.Text); len(vs) != 0 {
						t.Errorf("%s/%s/%s: generated text fails the check %+v:\n%s",
							st, lang, ch, vs, p.Text)
					}
				}
			}
		}
	}
}

// Compose runs the check itself, so a caller cannot forget to.
//
// The authority name is the vector that matters: it is free text from the
// database and it is interpolated straight into the opening sentence.
func TestComposeRefusesToReturnTextThatFailsTheCheck(t *testing.T) {
	f := lodgedFacts()
	f.Authority = "BMC, a corrupt body"
	if _, err := Compose(f, EN, WhatsApp); err == nil {
		t.Error("Compose returned a post that fails its own tone check")
	} else if !strings.Contains(err.Error(), "wrongdoing") {
		t.Errorf("error does not name the violation: %v", err)
	}
}

// A subcategory the vocabulary does not know is replaced by the category's
// word rather than printed raw, so free text from the classifier cannot reach
// a public post at all — and the substitution is recorded, because a silent
// one would read as a finished sentence.
func TestAnUnknownSubcategoryIsReplacedRatherThanPrinted(t *testing.T) {
	f := lodgedFacts()
	f.Subcategory = "pothole caused by a corrupt contractor"
	p, err := Compose(f, EN, WhatsApp)
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	if strings.Contains(p.Text, "corrupt") {
		t.Errorf("classifier free text reached the post:\n%s", p.Text)
	}
	if !strings.Contains(p.Text, "Road defect") {
		t.Errorf("the category word did not stand in for it:\n%s", p.Text)
	}
	if len(p.Fallbacks) == 0 {
		t.Error("the substitution was not recorded in Fallbacks")
	}
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
