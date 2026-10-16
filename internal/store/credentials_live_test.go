package store_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/auth"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// A test account, removed afterwards. The address is obviously not a person's.
func liveAccount(t *testing.T) (string, *store.DB) {
	t.Helper()
	ctx, pool, db := liveDB(t)

	email := fmt.Sprintf("credential-test+%d@invalid.test", time.Now().UnixNano())
	hash, err := auth.HashPassword("a-reasonable-passphrase")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	id, err := db.CreateEmailAccount(ctx, email, hash, "Credential Test")
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(ctx, `DELETE FROM accounts WHERE id = $1::uuid`, id); err != nil {
			t.Logf("could not remove the test account %s: %v", id, err)
		}
	})
	return id, db
}

// The invariant the whole design rests on: a one-time credential works once.
// Everything else — the expiry, the purpose, the single answer for four
// different failures — is a consequence.
func TestLiveAOneTimeCredentialIsSpentExactlyOnce(t *testing.T) {
	id, db := liveAccount(t)
	ctx, _, _ := liveDB(t)

	ot, err := auth.NewOneTime()
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if err := db.CreateAuthToken(ctx, id, auth.PasswordReset, ot.Hash,
		time.Now().UTC().Add(auth.PasswordReset.TTL())); err != nil {
		t.Fatalf("create credential: %v", err)
	}

	got, err := db.ConsumeAuthToken(ctx, auth.PasswordReset, ot.Hash)
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if got != id {
		t.Errorf("the credential named %q, want %q", got, id)
	}
	if _, err := db.ConsumeAuthToken(ctx, auth.PasswordReset, ot.Hash); !errors.Is(err, store.ErrNoCredential) {
		t.Errorf("spending it twice gave %v, want ErrNoCredential", err)
	}
}

// Four different failures, one answer. Telling them apart would tell somebody
// holding a stolen link whether it was ever valid.
func TestLiveACredentialIsRefusedForEveryReasonIdentically(t *testing.T) {
	id, db := liveAccount(t)
	ctx, _, _ := liveDB(t)

	expired, err := auth.NewOneTime()
	if err != nil {
		t.Fatal(err)
	}
	// Expiry is enforced by the query, so the row has to be written with one
	// in the past. The CHECK constraint forbids expires_at <= created_at, so
	// it is created and then moved.
	if err := db.CreateAuthToken(ctx, id, auth.PasswordReset, expired.Hash,
		time.Now().UTC().Add(time.Minute)); err != nil {
		t.Fatalf("create credential: %v", err)
	}
	_, pool, _ := liveDB(t)
	if _, err := pool.Exec(ctx, `
		UPDATE auth_tokens SET created_at = now() - interval '2 hours',
		                       expires_at = now() - interval '1 hour'
		 WHERE token_hash = $1`, expired.Hash); err != nil {
		t.Fatalf("age the credential: %v", err)
	}

	verify, err := auth.NewOneTime()
	if err != nil {
		t.Fatal(err)
	}
	if err := db.CreateAuthToken(ctx, id, auth.EmailVerification, verify.Hash,
		time.Now().UTC().Add(auth.EmailVerification.TTL())); err != nil {
		t.Fatalf("create credential: %v", err)
	}

	forged, err := auth.NewOneTime()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name    string
		purpose auth.Purpose
		hash    []byte
	}{
		{"a forged credential", auth.PasswordReset, forged.Hash},
		{"an expired credential", auth.PasswordReset, expired.Hash},
		{"a credential minted for another purpose", auth.PasswordReset, verify.Hash},
		{"a purpose the platform does not issue", auth.Purpose("anything"), verify.Hash},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := db.ConsumeAuthToken(ctx, c.purpose, c.hash); !errors.Is(err, store.ErrNoCredential) {
				t.Errorf("got %v, want ErrNoCredential", err)
			}
		})
	}
}

// A reset exists because a password may have been stolen, so it must end
// every session and every other pending reset in the same breath.
func TestLiveSettingAPasswordEndsSessionsAndOtherPendingResets(t *testing.T) {
	id, db := liveAccount(t)
	ctx, _, _ := liveDB(t)

	before, err := db.TokenVersion(ctx, id)
	if err != nil {
		t.Fatalf("token version: %v", err)
	}

	first, _ := auth.NewOneTime()
	second, _ := auth.NewOneTime()
	for _, ot := range []auth.OneTime{first, second} {
		if err := db.CreateAuthToken(ctx, id, auth.PasswordReset, ot.Hash,
			time.Now().UTC().Add(auth.PasswordReset.TTL())); err != nil {
			t.Fatalf("create credential: %v", err)
		}
	}

	if _, err := db.ConsumeAuthToken(ctx, auth.PasswordReset, first.Hash); err != nil {
		t.Fatalf("consume: %v", err)
	}
	hash, err := auth.HashPassword("a-different-passphrase")
	if err != nil {
		t.Fatal(err)
	}
	after, err := db.SetPassword(ctx, id, hash)
	if err != nil {
		t.Fatalf("set password: %v", err)
	}
	if after <= before {
		t.Errorf("token version went %d -> %d; every session should have ended",
			before, after)
	}

	// The second link, which this person did not ask for, is dead too.
	if _, err := db.ConsumeAuthToken(ctx, auth.PasswordReset, second.Hash); !errors.Is(err, store.ErrNoCredential) {
		t.Errorf("a second reset link survived the reset: %v", err)
	}

	// And the new password is the one that works.
	acct, err := db.FindAccountByEmail(ctx, accountEmail(t, id))
	if err != nil {
		t.Fatalf("find account: %v", err)
	}
	if !auth.VerifyPassword(acct.PasswordHash, "a-different-passphrase") {
		t.Error("the new password does not verify")
	}
	if auth.VerifyPassword(acct.PasswordHash, "a-reasonable-passphrase") {
		t.Error("the old password still verifies")
	}
	if acct.TokenVersion != after {
		t.Errorf("the account reads version %d, SetPassword returned %d",
			acct.TokenVersion, after)
	}
}

// Signing out everywhere ends sessions and leaves the password alone. The
// difference matters: one is a person who lost a phone, the other is a person
// whose password was stolen.
func TestLiveRevokingSessionsIsNotALockout(t *testing.T) {
	id, db := liveAccount(t)
	ctx, _, _ := liveDB(t)

	before, err := db.TokenVersion(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	after, err := db.RevokeSessions(ctx, id)
	if err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if after != before+1 {
		t.Errorf("version went %d -> %d, want +1", before, after)
	}
	acct, err := db.FindAccountByEmail(ctx, accountEmail(t, id))
	if err != nil {
		t.Fatalf("find account: %v", err)
	}
	if !auth.VerifyPassword(acct.PasswordHash, "a-reasonable-passphrase") {
		t.Error("revoking sessions changed the password")
	}

	if _, err := db.RevokeSessions(ctx, "00000000-0000-0000-0000-000000000000"); !errors.Is(err, store.ErrNoAccount) {
		t.Errorf("revoking for an unknown account gave %v, want ErrNoAccount", err)
	}
	if _, err := db.TokenVersion(ctx, "00000000-0000-0000-0000-000000000000"); !errors.Is(err, store.ErrNoAccount) {
		t.Errorf("reading the version of an unknown account gave %v, want ErrNoAccount", err)
	}
}

// Verification is idempotent and does not move the date: the first proof is
// the one that happened.
func TestLiveMarkingAnAddressVerifiedKeepsTheFirstDate(t *testing.T) {
	id, db := liveAccount(t)
	ctx, pool, _ := liveDB(t)

	info, err := db.GetAccount(ctx, id)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if info.EmailVerified {
		t.Fatal("a new password account is already verified")
	}

	if err := db.MarkEmailVerified(ctx, id); err != nil {
		t.Fatalf("mark verified: %v", err)
	}
	var first time.Time
	if err := pool.QueryRow(ctx,
		`SELECT email_verified_at FROM accounts WHERE id = $1::uuid`, id).Scan(&first); err != nil {
		t.Fatalf("read date: %v", err)
	}
	if err := db.MarkEmailVerified(ctx, id); err != nil {
		t.Fatalf("mark verified again: %v", err)
	}
	var second time.Time
	if err := pool.QueryRow(ctx,
		`SELECT email_verified_at FROM accounts WHERE id = $1::uuid`, id).Scan(&second); err != nil {
		t.Fatalf("read date: %v", err)
	}
	if !second.Equal(first) {
		t.Errorf("the verification date moved from %v to %v", first, second)
	}

	info, err = db.GetAccount(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if !info.EmailVerified {
		t.Error("the account does not read as verified")
	}
}

// accountEmail reads the address back, because the test account's address is
// generated and the store normalises it.
func accountEmail(t *testing.T, id string) string {
	t.Helper()
	ctx, pool, _ := liveDB(t)
	var email string
	if err := pool.QueryRow(ctx,
		`SELECT email FROM accounts WHERE id = $1::uuid`, id).Scan(&email); err != nil {
		t.Fatalf("read email: %v", err)
	}
	return email
}
