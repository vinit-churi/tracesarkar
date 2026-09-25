package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/vinit-churi/tracesarkar/internal/auth"
)

// Errors callers are expected to handle rather than surface verbatim.
var (
	ErrEmailTaken = errors.New("that email address is already registered")
	ErrNoAccount  = errors.New("no such account")
)

// AccountAuth is what a sign-in check needs.
type AccountAuth struct {
	ID           string
	Email        string
	PasswordHash []byte
	DisplayName  string
}

// AccountInfo is what a client may see about itself. The tags matter: this
// struct is serialised straight to clients.
type AccountInfo struct {
	ID            string   `json:"id"`
	Email         string   `json:"email"`
	DisplayName   string   `json:"name"`
	AuthProviders []string `json:"auth_providers"`
}

// CreateEmailAccount registers a password account.
func (d *DB) CreateEmailAccount(ctx context.Context, email string, passwordHash []byte, displayName string) (string, error) {
	normalised := auth.NormaliseEmail(email)
	var id string
	err := d.pool.QueryRow(ctx, `
		INSERT INTO accounts (id, email, email_normalised, password_hash, display_name,
		                      display_handle, auth_providers)
		VALUES (gen_random_uuid(), $1, $2, $3, NULLIF($4,''), $2, ARRAY['password'])
		RETURNING id::text`,
		email, normalised, passwordHash, displayName).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return "", ErrEmailTaken
		}
		return "", fmt.Errorf("create account: %w", err)
	}
	return id, nil
}

// FindAccountByEmail returns the account for an address, or ErrNoAccount.
func (d *DB) FindAccountByEmail(ctx context.Context, email string) (AccountAuth, error) {
	var a AccountAuth
	var displayName *string
	err := d.pool.QueryRow(ctx, `
		SELECT id::text, COALESCE(email,''), password_hash, display_name
		FROM accounts WHERE email_normalised = $1`, auth.NormaliseEmail(email)).
		Scan(&a.ID, &a.Email, &a.PasswordHash, &displayName)
	if errors.Is(err, pgx.ErrNoRows) {
		return AccountAuth{}, ErrNoAccount
	}
	if err != nil {
		return AccountAuth{}, fmt.Errorf("find account: %w", err)
	}
	a.DisplayName = derefString(displayName)
	return a, nil
}

// UpsertGoogleAccount finds or creates the account behind a Google identity.
//
// The lookup order matters: the Google subject first, because it is stable and
// the address on it can change; then the email address, so that someone who
// registered with a password and later signs in with Google keeps one account
// rather than silently acquiring a second.
func (d *DB) UpsertGoogleAccount(ctx context.Context, identity auth.Identity) (id string, created bool, err error) {
	email := auth.NormaliseEmail(identity.Email)

	err = d.pool.QueryRow(ctx,
		`SELECT id::text FROM accounts WHERE google_sub = $1`, identity.Subject).Scan(&id)
	if err == nil {
		return id, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", false, fmt.Errorf("find google account: %w", err)
	}

	if email != "" {
		err = d.pool.QueryRow(ctx, `
			UPDATE accounts
			SET google_sub = $1,
			    display_name = COALESCE(display_name, NULLIF($3,'')),
			    auth_providers = (
			      SELECT array_agg(DISTINCT p) FROM unnest(auth_providers || ARRAY['google']) AS p)
			WHERE email_normalised = $2 AND google_sub IS NULL
			RETURNING id::text`, identity.Subject, email, identity.Name).Scan(&id)
		if err == nil {
			return id, false, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return "", false, fmt.Errorf("link google account: %w", err)
		}
	}

	err = d.pool.QueryRow(ctx, `
		INSERT INTO accounts (id, email, email_normalised, google_sub, display_name,
		                      display_handle, auth_providers)
		VALUES (gen_random_uuid(), NULLIF($1,''), NULLIF($2,''), $3, NULLIF($4,''),
		        NULLIF($2,''), ARRAY['google'])
		RETURNING id::text`,
		identity.Email, email, identity.Subject, identity.Name).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return "", false, ErrEmailTaken
		}
		return "", false, fmt.Errorf("create google account: %w", err)
	}
	return id, true, nil
}

// GetAccount returns what a client may see about itself.
func (d *DB) GetAccount(ctx context.Context, id string) (AccountInfo, error) {
	var info AccountInfo
	var email, displayName *string
	err := d.pool.QueryRow(ctx, `
		SELECT id::text, email, display_name, auth_providers
		FROM accounts WHERE id = $1::uuid`, id).
		Scan(&info.ID, &email, &displayName, &info.AuthProviders)
	if errors.Is(err, pgx.ErrNoRows) {
		return AccountInfo{}, ErrNoAccount
	}
	if err != nil {
		return AccountInfo{}, fmt.Errorf("get account: %w", err)
	}
	info.Email = auth.NormaliseEmail(derefString(email))
	info.DisplayName = derefString(displayName)
	return info, nil
}
