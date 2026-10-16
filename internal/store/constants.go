package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Constant is one published figure the platform is allowed to state — a
// deadline, a fee, a penalty — with the document that creates it.
//
// Hard rule 6 keeps these out of Go source. This type is how code reads one
// back: never a literal, and never without the citation, because a number
// shown on a public surface with nothing behind it is the thing the rule
// exists to prevent.
type Constant struct {
	Key string
	// Value is the raw JSON scalar as text — "48", "600000". Callers parse it
	// for the unit they expect rather than this layer guessing between hours,
	// days and paise.
	Value         string
	Citation      string
	ShortCitation string
	CitationURL   string
	Verification  string
	EffectiveFrom time.Time
}

// LegalConstant returns the constant in force today for a key.
//
// Not found is not an error. Most keys are absent — the platform holds five
// categories and a handful of published clocks — and the honest answer on a
// surface is to say nothing rather than to borrow a number from elsewhere.
func (d *DB) LegalConstant(ctx context.Context, key string) (Constant, bool, error) {
	c := Constant{Key: strings.TrimSpace(key)}
	err := d.pool.QueryRow(ctx, `
		SELECT value::text, citation, COALESCE(short_citation, ''),
		       COALESCE(citation_url, ''), verification, effective_from
		  FROM legal_constants
		 WHERE key = $1 AND effective_from <= now()::date
		   AND (effective_to IS NULL OR effective_to >= now()::date)
		 ORDER BY effective_from DESC
		 LIMIT 1`, c.Key).
		Scan(&c.Value, &c.Citation, &c.ShortCitation, &c.CitationURL,
			&c.Verification, &c.EffectiveFrom)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Constant{}, false, nil
		}
		return Constant{}, false, fmt.Errorf("legal constant %s: %w", c.Key, err)
	}
	// Stored as a JSON scalar, so a number arrives bare and a string quoted.
	c.Value = strings.Trim(c.Value, `"`)
	return c, true, nil
}
