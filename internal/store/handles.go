package store

import (
	"context"
	"fmt"

	"github.com/vinit-churi/tracesarkar/internal/share"
)

// Handles returns the accounts a generated post may tag for this authority.
//
// Three filters are applied here rather than in the caller, because a caller
// that forgets one tags a stranger:
//
//   - in date. A row past `verify_by` is not served. Handles get renamed,
//     merged and abandoned, and the registry is only as good as its last
//     check.
//   - in force. `effective_from` in the future is not yet the answer.
//   - scoped. A ward row wins over an authority-wide row for that ward, and an
//     authority-wide row still applies where no ward row exists.
//
// Nothing here decides whether an escalation handle may be used. That depends
// on the state of a deadline, which is the share kit's judgement to make
// (share.Compose), not the database's.
func (d *DB) Handles(ctx context.Context, authority, ward string) ([]share.Handle, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT platform, handle, kind, source_ref, retrieved_at, first_hand
		  FROM authority_handles
		 WHERE authority = $1
		   AND (ward IS NULL OR ward = $2)
		   AND effective_from <= CURRENT_DATE
		   AND verify_by      >= CURRENT_DATE
		 ORDER BY (ward IS NOT NULL) DESC, platform, handle`, authority, ward)
	if err != nil {
		return nil, fmt.Errorf("list handles for %s: %w", authority, err)
	}
	defer rows.Close()

	var out []share.Handle
	for rows.Next() {
		var h share.Handle
		var kind string
		if err := rows.Scan(&h.Platform, &h.Handle, &kind, &h.SourceRef,
			&h.RetrievedAt, &h.FirstHand); err != nil {
			return nil, fmt.Errorf("scan handle: %w", err)
		}
		h.Kind = share.HandleKind(kind)
		out = append(out, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read handles for %s: %w", authority, err)
	}
	return out, nil
}
