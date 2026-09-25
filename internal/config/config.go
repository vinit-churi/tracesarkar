// Package config loads runtime configuration from the environment and an
// optional .env file. Secrets are never logged and never appear in errors.
package config

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// R2 holds the S3-compatible object storage settings for the archive.
type R2 struct {
	Endpoint  string // https://<account>.r2.cloudflarestorage.com
	Bucket    string
	Region    string
	AccessKey string
	Secret    string
}

// Postgres holds the database connection settings.
type Postgres struct {
	URL    string
	CAPath string // optional CA bundle for servers with a private root
	// CAPem is the same bundle inline. Some platforms — Dokploy, Heroku-style
	// hosts — have no way to mount a file, only environment variables, so the
	// certificate has to travel as one. CAFile turns it back into a path.
	CAPem string
}

// CAFile returns a path the Postgres driver can open, materialising an inline
// PEM into a temporary file if that is all we were given. An explicit path
// wins: a mounted file is the more deliberate configuration of the two.
func (p Postgres) CAFile() (string, error) {
	if strings.TrimSpace(p.CAPath) != "" {
		return p.CAPath, nil
	}
	pem := strings.TrimSpace(p.CAPem)
	if pem == "" {
		return "", nil
	}
	// A .env file is a poor container for a multi-line value and every platform
	// parses one slightly differently, so the PEM may arrive base64-encoded.
	if !strings.Contains(pem, "BEGIN CERTIFICATE") {
		if decoded, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(pem), "")); err == nil {
			pem = strings.TrimSpace(string(decoded))
		}
	}
	// Fail here, with a name, rather than as an opaque handshake error later.
	if !strings.Contains(pem, "BEGIN CERTIFICATE") {
		return "", errors.New("POSTGRES_CA_PEM is neither a PEM certificate nor base64 of one")
	}

	f, err := os.CreateTemp("", "tracesarkar-ca-*.pem")
	if err != nil {
		return "", fmt.Errorf("materialise postgres CA: %w", err)
	}
	defer f.Close()
	if err := f.Chmod(0o600); err != nil {
		return "", fmt.Errorf("materialise postgres CA: %w", err)
	}
	if _, err := f.WriteString(pem + "\n"); err != nil {
		return "", fmt.Errorf("materialise postgres CA: %w", err)
	}
	return f.Name(), nil
}

// Config is everything a Phase 0 binary needs to run.
type Config struct {
	R2       R2
	Postgres Postgres
	// NotifyWebhook is optional; when empty, alerts are logged only.
	NotifyWebhook string
	// Tier is the exposure tier this process runs at (D044).
	Tier string
}

const defaultTier = "personal"

// Load reads .env (if present) and then the process environment, which wins.
func Load(envFile string) (Config, error) {
	values, err := parseEnvFile(envFile)
	if err != nil {
		return Config{}, err
	}
	fromFile := map[string]bool{}
	for k := range values {
		fromFile[k] = true
	}
	for _, key := range []string{
		"R2_BUCKET_URL", "R2_BUCKET_NAME", "R2_ACCESS_KEY", "R2_SECRET_ACCESS_KEY",
		"POSTGRESQL_CONNECTION", "POSTGRES_CA_PATH", "POSTGRES_CA_PEM",
		"NOTIFY_WEBHOOK_URL", "TRACESARKAR_TIER",
	} {
		if v, ok := os.LookupEnv(key); ok && v != "" {
			values[key] = v
			fromFile[key] = false
		}
	}

	cfg, err := loadFrom(values)
	if err != nil {
		return Config{}, err
	}
	// A CA path written in .env is relative to that file, not to the working
	// directory of whichever binary or test is running.
	if cfg.Postgres.CAPath != "" && !filepath.IsAbs(cfg.Postgres.CAPath) && fromFile["POSTGRES_CA_PATH"] {
		cfg.Postgres.CAPath = filepath.Join(filepath.Dir(envFile), cfg.Postgres.CAPath)
	}
	return cfg, nil
}

func loadFrom(values map[string]string) (Config, error) {
	var missing []string
	need := func(key string) string {
		v := strings.TrimSpace(values[key])
		if v == "" {
			missing = append(missing, key)
		}
		return v
	}

	bucketURL := need("R2_BUCKET_URL")
	accessKey := need("R2_ACCESS_KEY")
	secret := need("R2_SECRET_ACCESS_KEY")
	pgURL := need("POSTGRESQL_CONNECTION")

	if len(missing) > 0 {
		sort.Strings(missing)
		// Only key names are reported. Values may be secrets.
		return Config{}, fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}

	endpoint, bucketFromURL, err := splitBucketURL(bucketURL)
	if err != nil {
		return Config{}, err
	}
	bucket := strings.TrimSpace(values["R2_BUCKET_NAME"])
	if bucket == "" {
		bucket = bucketFromURL
	}
	if bucket == "" {
		return Config{}, errors.New("missing required configuration: R2_BUCKET_NAME")
	}

	tier := strings.TrimSpace(values["TRACESARKAR_TIER"])
	if tier == "" {
		tier = defaultTier
	}

	return Config{
		R2: R2{
			Endpoint:  endpoint,
			Bucket:    bucket,
			Region:    "auto", // R2 signs with "auto"
			AccessKey: accessKey,
			Secret:    secret,
		},
		Postgres: Postgres{
			URL:    pgURL,
			CAPath: strings.TrimSpace(values["POSTGRES_CA_PATH"]),
			CAPem:  strings.TrimSpace(values["POSTGRES_CA_PEM"]),
		},
		NotifyWebhook: strings.TrimSpace(values["NOTIFY_WEBHOOK_URL"]),
		Tier:          tier,
	}, nil
}

// splitBucketURL turns "https://acc.r2.cloudflarestorage.com/bucket" into its
// endpoint and bucket parts. A URL with no path yields an empty bucket.
func splitBucketURL(raw string) (endpoint, bucket string, err error) {
	trimmed := strings.TrimSuffix(strings.TrimSpace(raw), "/")
	scheme := ""
	rest := trimmed
	for _, s := range []string{"https://", "http://"} {
		if strings.HasPrefix(trimmed, s) {
			scheme, rest = s, strings.TrimPrefix(trimmed, s)
			break
		}
	}
	if scheme == "" {
		return "", "", fmt.Errorf("R2_BUCKET_URL must start with https:// or http://")
	}
	host, path, _ := strings.Cut(rest, "/")
	if host == "" {
		return "", "", fmt.Errorf("R2_BUCKET_URL has no host")
	}
	return scheme + host, path, nil
}

// parseEnvFile reads KEY=VALUE lines. A missing file is not an error.
func parseEnvFile(path string) (map[string]string, error) {
	values := map[string]string{}
	if path == "" {
		return values, nil
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return values, nil
		}
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return values, nil
}
