package store

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationsAreNumberedOrderedAndUnique(t *testing.T) {
	files, err := migrationFiles()
	if err != nil {
		t.Fatalf("migrationFiles: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no migrations found")
	}

	seen := map[int]string{}
	last := 0
	for _, f := range files {
		if other, dup := seen[f.Version]; dup {
			t.Errorf("version %d used twice: %s and %s", f.Version, other, f.Name)
		}
		seen[f.Version] = f.Name
		if f.Version <= last {
			t.Errorf("%s is out of order: %d after %d", f.Name, f.Version, last)
		}
		last = f.Version
		if strings.TrimSpace(f.SQL) == "" {
			t.Errorf("%s is empty", f.Name)
		}
	}
}

func TestParseMigrationName(t *testing.T) {
	tests := []struct {
		in      string
		version int
		ok      bool
	}{
		{"0001_phase0_instruments.sql", 1, true},
		{"0012_something.sql", 12, true},
		{"not-a-migration.sql", 0, false},
		{"0001.sql", 1, true},
		{"README.md", 0, false},
	}
	for _, tt := range tests {
		got, ok := parseMigrationVersion(tt.in)
		if ok != tt.ok || got != tt.version {
			t.Errorf("%s: got (%d,%v), want (%d,%v)", tt.in, got, ok, tt.version, tt.ok)
		}
	}
}

func TestSSLSettingsAddCARootWhenProvided(t *testing.T) {
	ca := filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(ca, []byte("-----BEGIN CERTIFICATE-----\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := applyTLS("postgres://u:p@h:5432/db?sslmode=require", ca)
	if err != nil {
		t.Fatalf("applyTLS: %v", err)
	}
	if !strings.Contains(got, url.QueryEscape(ca)) && !strings.Contains(got, ca) {
		t.Errorf("CA path missing from %q", got)
	}
	if !strings.Contains(got, "sslmode=verify-full") {
		t.Errorf("a CA should upgrade verification: %q", got)
	}
}

func TestSSLSettingsLeftAloneWithoutCA(t *testing.T) {
	const in = "postgres://u:p@h:5432/db?sslmode=require"
	got, err := applyTLS(in, "")
	if err != nil {
		t.Fatalf("applyTLS: %v", err)
	}
	if got != in {
		t.Errorf("got %q, want it unchanged", got)
	}
}

func TestApplyTLSFailsWhenCAFileIsMissing(t *testing.T) {
	_, err := applyTLS("postgres://u:p@h:5432/db", filepath.Join(t.TempDir(), "absent.pem"))
	if err == nil {
		t.Fatal("a configured CA that does not exist must be an error, not a silent downgrade")
	}
}

func TestRedactRemovesPasswordFromErrors(t *testing.T) {
	err := redact(errorString("failed to connect to postgres://avnadmin:sup3rs3cret@host:19839/db"))
	if strings.Contains(err.Error(), "sup3rs3cret") {
		t.Fatalf("password survived redaction: %v", err)
	}
}

type errorString string

func (e errorString) Error() string { return string(e) }
