package store_test

import (
	"os"
	"strings"
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/evidence"
	"github.com/vinit-churi/tracesarkar/internal/share"
)

// The pack's SQL is several joins across six tables, and a typo in one of them
// only shows up against the real schema. Read-only: it compiles a record that
// already exists and asserts the rules, and writes nothing.
func TestLiveAnEvidencePackCompilesFromARealIssue(t *testing.T) {
	ctx, _, db := liveDB(t)

	issueID := os.Getenv("EVIDENCE_ISSUE_ID")
	accountID := os.Getenv("EVIDENCE_ACCOUNT_ID")
	if issueID == "" || accountID == "" {
		t.Skip("set EVIDENCE_ISSUE_ID and EVIDENCE_ACCOUNT_ID to compile a real pack")
	}

	r, photos, ok, err := db.EvidenceFor(ctx, issueID, accountID)
	if err != nil {
		t.Fatalf("EvidenceFor: %v", err)
	}
	if !ok {
		t.Fatal("the reporter's own issue was not found")
	}
	if r.IssueID != issueID {
		t.Errorf("compiled issue %q, asked for %q", r.IssueID, issueID)
	}
	if r.Category == "" || r.FirstSeen.IsZero() {
		t.Errorf("the record is missing the facts a cover needs: %+v", r)
	}

	// Hard rule 3, over a record built from real data rather than a fixture.
	for _, line := range evidence.Lines(r) {
		if vs := share.Check(line); len(vs) != 0 {
			t.Errorf("a real record fails the tone check: %q %+v", line, vs)
		}
	}
	// Hard rule 2: nothing is named without the document behind it.
	if r.Desk != nil && strings.TrimSpace(r.Desk.SourceRef) == "" {
		t.Error("a desk was compiled with no source reference")
	}
	if r.Contract != nil && strings.TrimSpace(r.Contract.SourceRef) == "" {
		t.Error("a contract was compiled with no source reference")
	}
	// Only redacted copies the gate published.
	for _, p := range photos {
		if !p.Published {
			t.Errorf("a held photograph reached the pack: %s", p.Key)
		}
		if !strings.HasPrefix(p.Key, "derivative/") {
			t.Errorf("a photograph in the pack is not a redacted copy: %s", p.Key)
		}
	}

	out, err := evidence.Build(r, nil)
	if err != nil {
		t.Fatalf("Build from a real record: %v", err)
	}
	if len(out) < 1000 {
		t.Errorf("the pack is %d bytes, which is not a document", len(out))
	}
	t.Logf("compiled issue %s: %d photographs, stamp %s, %d bytes",
		r.IssueID, len(photos), evidence.Stamp(r), len(out))
	if path := os.Getenv("EVIDENCE_OUT"); path != "" {
		if err := os.WriteFile(path, out, 0o600); err != nil {
			t.Fatalf("write pack: %v", err)
		}
	}

	// A stranger gets nothing, and is told the issue does not exist.
	if _, _, found, err := db.EvidenceFor(ctx,
		issueID, "11111111-1111-4111-8111-111111111111"); err != nil || found {
		t.Errorf("a stranger compiled a pack (found=%v, err=%v)", found, err)
	}
}
