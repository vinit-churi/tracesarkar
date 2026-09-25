package config

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseEnvFile(t *testing.T) {
	tests := []struct {
		name string
		body string
		want map[string]string
	}{
		{
			name: "plain assignments",
			body: "A=1\nB=two\n",
			want: map[string]string{"A": "1", "B": "two"},
		},
		{
			name: "comments and blank lines are ignored",
			body: "# a comment\n\nA=1\n  # indented comment\n",
			want: map[string]string{"A": "1"},
		},
		{
			name: "quotes are stripped and values keep inner characters",
			body: "A=\"pg://u:p@h/db?sslmode=require\"\nB='x=y'\n",
			want: map[string]string{"A": "pg://u:p@h/db?sslmode=require", "B": "x=y"},
		},
		{
			name: "surrounding whitespace is trimmed",
			body: "  A = 1 \n",
			want: map[string]string{"A": "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), ".env")
			if err := os.WriteFile(path, []byte(tt.body), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := parseEnvFile(path)
			if err != nil {
				t.Fatalf("parseEnvFile: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d keys, want %d: %v", len(got), len(tt.want), got)
			}
			for k, want := range tt.want {
				if got[k] != want {
					t.Errorf("key %q: got %q, want %q", k, got[k], want)
				}
			}
		})
	}
}

func TestParseEnvFileMissingIsNotAnError(t *testing.T) {
	got, err := parseEnvFile(filepath.Join(t.TempDir(), "absent"))
	if err != nil {
		t.Fatalf("missing file should not error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no values, got %v", got)
	}
}

func TestLoadDerivesS3EndpointFromBucketURL(t *testing.T) {
	env := map[string]string{
		"R2_BUCKET_URL":         "https://acc123.r2.cloudflarestorage.com/tracesarkar",
		"R2_ACCESS_KEY":         "ak",
		"R2_SECRET_ACCESS_KEY":  "sk",
		"R2_BUCKET_NAME":        "tracesarkar",
		"POSTGRESQL_CONNECTION": "postgres://u:p@h:1/db?sslmode=require",
	}
	cfg, err := loadFrom(env)
	if err != nil {
		t.Fatalf("loadFrom: %v", err)
	}
	if cfg.R2.Endpoint != "https://acc123.r2.cloudflarestorage.com" {
		t.Errorf("endpoint: got %q", cfg.R2.Endpoint)
	}
	if cfg.R2.Bucket != "tracesarkar" {
		t.Errorf("bucket: got %q", cfg.R2.Bucket)
	}
	if cfg.R2.Region != "auto" {
		t.Errorf("region: got %q, want auto", cfg.R2.Region)
	}
	if cfg.Postgres.URL != env["POSTGRESQL_CONNECTION"] {
		t.Errorf("postgres url: got %q", cfg.Postgres.URL)
	}
}

func TestLoadReportsEveryMissingRequiredKey(t *testing.T) {
	_, err := loadFrom(map[string]string{"R2_ACCESS_KEY": "ak"})
	if err == nil {
		t.Fatal("expected an error when required keys are missing")
	}
	for _, want := range []string{"R2_BUCKET_URL", "R2_SECRET_ACCESS_KEY", "POSTGRESQL_CONNECTION"} {
		if !contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
	if contains(err.Error(), "R2_ACCESS_KEY") {
		t.Errorf("error should not mention the key that was present: %q", err)
	}
}

func TestLoadNeverPutsSecretsInErrors(t *testing.T) {
	_, err := loadFrom(map[string]string{
		"R2_ACCESS_KEY":        "AKIA-super-secret",
		"R2_SECRET_ACCESS_KEY": "shhh-very-secret",
	})
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, secret := range []string{"AKIA-super-secret", "shhh-very-secret"} {
		if contains(err.Error(), secret) {
			t.Fatalf("error leaked a secret value: %q", err)
		}
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}

func TestLoadResolvesRelativeCAPathAgainstTheEnvFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	body := "R2_BUCKET_URL=https://acc.r2.cloudflarestorage.com/b\n" +
		"R2_BUCKET_NAME=b\nR2_ACCESS_KEY=ak\nR2_SECRET_ACCESS_KEY=sk\n" +
		"POSTGRESQL_CONNECTION=postgres://u:p@h:1/db\n" +
		"POSTGRES_CA_PATH=./ca.pem\n"
	if err := os.WriteFile(envPath, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(envPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	want := filepath.Join(dir, "ca.pem")
	if cfg.Postgres.CAPath != want {
		t.Errorf("CA path: got %q, want %q", cfg.Postgres.CAPath, want)
	}
}

func TestLoadLeavesAbsoluteCAPathAlone(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	body := "R2_BUCKET_URL=https://acc.r2.cloudflarestorage.com/b\n" +
		"R2_BUCKET_NAME=b\nR2_ACCESS_KEY=ak\nR2_SECRET_ACCESS_KEY=sk\n" +
		"POSTGRESQL_CONNECTION=postgres://u:p@h:1/db\n" +
		"POSTGRES_CA_PATH=/etc/ssl/ca.pem\n"
	if err := os.WriteFile(envPath, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(envPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Postgres.CAPath != "/etc/ssl/ca.pem" {
		t.Errorf("got %q", cfg.Postgres.CAPath)
	}
}

// A platform whose only secret channel is environment variables cannot mount a
// CA file. The PEM travels inline instead, and the process materialises it.
func TestCAFileMaterialisesAnInlinePEM(t *testing.T) {
	t.Setenv("R2_BUCKET_URL", "https://acc.r2.cloudflarestorage.com/bucket")
	t.Setenv("R2_ACCESS_KEY", "key")
	t.Setenv("R2_SECRET_ACCESS_KEY", "secret")
	t.Setenv("POSTGRESQL_CONNECTION", "postgres://u:p@host:5432/db")
	t.Setenv("POSTGRES_CA_PEM", "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n")

	cfg, err := Load(filepath.Join(t.TempDir(), "absent.env"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	path, err := cfg.Postgres.CAFile()
	if err != nil {
		t.Fatalf("CAFile: %v", err)
	}
	if path == "" {
		t.Fatal("an inline PEM must yield a path the driver can open")
	}
	t.Cleanup(func() { _ = os.Remove(path) })

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read materialised CA: %v", err)
	}
	if !strings.Contains(string(written), "BEGIN CERTIFICATE") {
		t.Errorf("the PEM must survive intact: %q", written)
	}

	// The CA is not secret, but the file is ours; nothing else on a shared host
	// has a reason to write to it.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("permissions: got %o, want 600", perm)
	}
}

func TestCAFilePrefersAnExplicitPathOverAnInlinePEM(t *testing.T) {
	dir := t.TempDir()
	explicit := filepath.Join(dir, "ca.pem")
	if err := os.WriteFile(explicit, []byte("on disk"), 0o600); err != nil {
		t.Fatal(err)
	}

	p := Postgres{CAPath: explicit, CAPem: "inline"}
	path, err := p.CAFile()
	if err != nil {
		t.Fatalf("CAFile: %v", err)
	}
	if path != explicit {
		t.Errorf("a mounted file wins: got %q, want %q", path, explicit)
	}
}

func TestCAFileIsEmptyWhenNeitherIsConfigured(t *testing.T) {
	path, err := Postgres{}.CAFile()
	if err != nil {
		t.Fatalf("CAFile: %v", err)
	}
	if path != "" {
		t.Errorf("no CA configured must stay no CA, got %q", path)
	}
}

func TestCAFileRefusesSomethingThatIsNotAPEM(t *testing.T) {
	// A truncated or mangled secret should fail here, with a clear message,
	// rather than as an opaque TLS handshake error later.
	if _, err := (Postgres{CAPem: "not a certificate"}).CAFile(); err == nil {
		t.Fatal("expected an error for a PEM without a certificate block")
	}
}

// A .env file is a poor container for a multi-line value, and every platform
// parses one slightly differently. Base64 sidesteps the question entirely.
func TestCAFileAcceptsABase64EncodedPEM(t *testing.T) {
	pem := "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n"
	encoded := base64.StdEncoding.EncodeToString([]byte(pem))

	path, err := (Postgres{CAPem: encoded}).CAFile()
	if err != nil {
		t.Fatalf("CAFile: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(path) })

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(written), "BEGIN CERTIFICATE") {
		t.Errorf("the decoded PEM must reach disk: %q", written)
	}
}

func TestCAFileStillRefusesBase64OfSomethingElse(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("just some text"))

	if _, err := (Postgres{CAPem: encoded}).CAFile(); err == nil {
		t.Fatal("base64 that decodes to a non-certificate must still be refused")
	}
}
