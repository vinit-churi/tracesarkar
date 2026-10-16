package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/vinit-churi/tracesarkar/internal/auth"
)

// ErrNoCredential is a one-time credential that cannot be used: unknown,
// expired, already spent, or minted for a different purpose.
//
// One error for all four on purpose. Telling a caller which would tell
// somebody holding a stolen link whether it was ever valid, and telling
// somebody guessing whether they are close.
var ErrNoCredential = errors.New("that link is not valid")

// CreateAuthToken records a one-time credential. Only the hash is stored.
func (d *DB) CreateAuthToken(ctx context.Context, accountID string, purpose auth.Purpose, hash []byte, expiresAt time.Time) error {
	if !purpose.Valid() {
		return fmt.Errorf("create credential: unknown purpose %q", purpose)
	}
	_, err := d.pool.Exec(ctx, `
		INSERT INTO auth_tokens (account_id, purpose, token_hash, expires_at)
		VALUES ($1::uuid, $2, $3, $4)`,
		accountID, string(purpose), hash, expiresAt.UTC())
	if err != nil {
		return fmt.Errorf("create credential: %w", err)
	}
	return nil
}

// ConsumeAuthToken spends a credential and returns whose account it was for.
//
// One statement, so the check and the spend cannot be separated. Two requests
// arriving with the same link at the same moment mean one of them finds the
// row already marked and gets ErrNoCredential, rather than both succeeding.
func (d *DB) ConsumeAuthToken(ctx context.Context, purpose auth.Purpose, hash []byte) (string, error) {
	if !purpose.Valid() {
		return "", ErrNoCredential
	}
	var accountID string
	err := d.pool.QueryRow(ctx, `
		UPDATE auth_tokens
		   SET used_at = now()
		 WHERE token_hash = $1
		   AND purpose = $2
		   AND used_at IS NULL
		   AND expires_at > now()
		RETURNING account_id::text`, hash, string(purpose)).Scan(&accountID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNoCredential
	}
	if err != nil {
		return "", fmt.Errorf("consume credential: %w", err)
	}
	return accountID, nil
}

// SetPassword replaces the password and ends every existing session.
//
// The two are one operation and must stay one. Somebody resetting a password
// because a password was stolen has to be left with no sessions they did not
// just create, and a reset that leaves the thief signed in is worse than no
// reset at all. Returns the new token version.
func (d *DB) SetPassword(ctx context.Context, accountID string, hash []byte) (int64, error) {
	var version int64
	err := d.pool.QueryRow(ctx, `
		UPDATE accounts
		   SET password_hash = $2,
		       token_version = token_version + 1,
		       auth_providers = (
		         SELECT array_agg(DISTINCT p)
		           FROM unnest(auth_providers || ARRAY['password']) AS p)
		 WHERE id = $1::uuid
		RETURNING token_version`, accountID, hash).Scan(&version)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNoAccount
	}
	if err != nil {
		return 0, fmt.Errorf("set password: %w", err)
	}
	// Every pending reset for this account is spent, including the ones this
	// person did not ask for. A second link still live after a reset is a
	// second chance for whoever requested it.
	if _, err := d.pool.Exec(ctx, `
		UPDATE auth_tokens SET used_at = now()
		 WHERE account_id = $1::uuid AND purpose = $2 AND used_at IS NULL`,
		accountID, string(auth.PasswordReset)); err != nil {
		return 0, fmt.Errorf("spend remaining reset links: %w", err)
	}
	return version, nil
}

// RevokeSessions ends every session for an account and returns the new
// version. The account keeps its password: this is "sign out everywhere", not
// a lockout.
func (d *DB) RevokeSessions(ctx context.Context, accountID string) (int64, error) {
	var version int64
	err := d.pool.QueryRow(ctx, `
		UPDATE accounts SET token_version = token_version + 1
		 WHERE id = $1::uuid
		RETURNING token_version`, accountID).Scan(&version)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNoAccount
	}
	if err != nil {
		return 0, fmt.Errorf("revoke sessions: %w", err)
	}
	return version, nil
}

// TokenVersion is the account's current version, read on every authenticated
// request so that a revoked token stops working within one request rather
// than at its own expiry.
func (d *DB) TokenVersion(ctx context.Context, accountID string) (int64, error) {
	var version int64
	err := d.pool.QueryRow(ctx,
		`SELECT token_version FROM accounts WHERE id = $1::uuid`, accountID).Scan(&version)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNoAccount
	}
	if err != nil {
		return 0, fmt.Errorf("read token version: %w", err)
	}
	return version, nil
}

// MarkEmailVerified records that the address reaches the person holding the
// account. Idempotent, and it does not move the date on a second
// confirmation: the first proof is the one that happened.
func (d *DB) MarkEmailVerified(ctx context.Context, accountID string) error {
	_, err := d.pool.Exec(ctx, `
		UPDATE accounts
		   SET email_verified_at = COALESCE(email_verified_at, now())
		 WHERE id = $1::uuid`, accountID)
	if err != nil {
		return fmt.Errorf("mark email verified: %w", err)
	}
	return nil
}

// DeleteSpentCredentials removes rows that can no longer be used. Kept for a
// week after they stop working so that a support question about a link that
// did not arrive can still be answered.
func (d *DB) DeleteSpentCredentials(ctx context.Context, olderThan time.Duration) (int64, error) {
	tag, err := d.pool.Exec(ctx, `
		DELETE FROM auth_tokens
		 WHERE (used_at IS NOT NULL OR expires_at < now())
		   AND created_at < now() - $1::interval`,
		fmt.Sprintf("%d seconds", int64(olderThan.Seconds())))
	if err != nil {
		return 0, fmt.Errorf("delete spent credentials: %w", err)
	}
	return tag.RowsAffected(), nil
}
