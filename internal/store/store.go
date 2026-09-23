// Package store owns the database connection and the forward-only migrations.
// SQL is written by hand; there is no ORM.
package store

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migration is one numbered, forward-only SQL file.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

var migrationNamePattern = regexp.MustCompile(`^(\d{1,6})(_[a-z0-9_]*)?\.sql$`)

func parseMigrationVersion(name string) (int, bool) {
	m := migrationNamePattern.FindStringSubmatch(name)
	if m == nil {
		return 0, false
	}
	v, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, false
	}
	return v, true
}

func migrationFiles() ([]Migration, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}
	var out []Migration
	for _, e := range entries {
		version, ok := parseMigrationVersion(e.Name())
		if !ok {
			continue
		}
		body, err := migrationsFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		out = append(out, Migration{Version: version, Name: e.Name(), SQL: string(body)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}

// applyTLS adds a CA root to the connection string when one is configured, and
// upgrades verification accordingly. Without a CA the string is untouched.
func applyTLS(connString, caPath string) (string, error) {
	if strings.TrimSpace(caPath) == "" {
		return connString, nil
	}
	if _, err := os.Stat(caPath); err != nil {
		return "", fmt.Errorf("postgres CA file: %w", err)
	}
	u, err := url.Parse(connString)
	if err != nil {
		return "", fmt.Errorf("parse postgres url: %w", err)
	}
	q := u.Query()
	q.Set("sslrootcert", caPath)
	q.Set("sslmode", "verify-full")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// Connect opens a pool. caPath may be empty.
func Connect(ctx context.Context, connString, caPath string) (*pgxpool.Pool, error) {
	withTLS, err := applyTLS(connString, caPath)
	if err != nil {
		return nil, err
	}
	cfg, err := pgxpool.ParseConfig(withTLS)
	if err != nil {
		// The connection string carries a password; never echo it.
		return nil, fmt.Errorf("parse postgres configuration: %w", redact(err))
	}
	cfg.MaxConns = 4
	cfg.MaxConnLifetime = 30 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", redact(err))
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", redact(err))
	}
	return pool, nil
}

// Migrate applies every migration that has not been applied yet, in order.
func Migrate(ctx context.Context, pool *pgxpool.Pool) ([]Migration, error) {
	files, err := migrationFiles()
	if err != nil {
		return nil, err
	}
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
		  version     int PRIMARY KEY,
		  name        text NOT NULL,
		  applied_at  timestamptz NOT NULL DEFAULT now()
		)`)
	if err != nil {
		return nil, fmt.Errorf("create schema_migrations: %w", err)
	}

	applied := map[int]bool{}
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("read schema_migrations: %w", err)
	}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan schema_migrations: %w", err)
		}
		applied[v] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read schema_migrations: %w", err)
	}

	var ran []Migration
	for _, m := range files {
		if applied[m.Version] {
			continue
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			return ran, fmt.Errorf("begin %s: %w", m.Name, err)
		}
		if _, err := tx.Exec(ctx, m.SQL); err != nil {
			_ = tx.Rollback(ctx)
			return ran, fmt.Errorf("apply %s: %w", m.Name, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, m.Version, m.Name); err != nil {
			_ = tx.Rollback(ctx)
			return ran, fmt.Errorf("record %s: %w", m.Name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return ran, fmt.Errorf("commit %s: %w", m.Name, err)
		}
		ran = append(ran, m)
	}
	return ran, nil
}

// redact strips anything that looks like a password from an error.
var passwordPattern = regexp.MustCompile(`(?i)(password=)[^\s&]+|(://[^:/@]+):[^@]+@`)

func redact(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s", passwordPattern.ReplaceAllString(err.Error(), "${1}${2}***"))
}
