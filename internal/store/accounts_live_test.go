package store_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/auth"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

func uniqueEmail() string {
	return fmt.Sprintf("test_%d@example.com", time.Now().UnixNano())
}

func TestLiveEmailAccountLifecycle(t *testing.T) {
	ctx, pool, db := liveDB(t)
	email := uniqueEmail()

	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM accounts WHERE email_normalised = $1`, auth.NormaliseEmail(email))
	})

	hash, err := auth.HashPassword("a-reasonable-passphrase")
	if err != nil {
		t.Fatal(err)
	}

	id, err := db.CreateEmailAccount(ctx, email, hash, "Test Person")
	if err != nil {
		t.Fatalf("CreateEmailAccount: %v", err)
	}

	// The same address in different case is the same person.
	_, err = db.CreateEmailAccount(ctx, "TEST_"+email[5:], hash, "Impostor")
	if !errors.Is(err, store.ErrEmailTaken) {
		t.Errorf("a duplicate email must be refused with ErrEmailTaken, got %v", err)
	}

	found, err := db.FindAccountByEmail(ctx, email)
	if err != nil {
		t.Fatalf("FindAccountByEmail: %v", err)
	}
	if found.ID != id {
		t.Errorf("found a different account: %s vs %s", found.ID, id)
	}
	if !auth.VerifyPassword(found.PasswordHash, "a-reasonable-passphrase") {
		t.Error("the stored hash does not verify the original password")
	}
	if auth.VerifyPassword(found.PasswordHash, "wrong") {
		t.Error("the stored hash verified a wrong password")
	}

	if _, err := db.FindAccountByEmail(ctx, "nobody_"+email); !errors.Is(err, store.ErrNoAccount) {
		t.Errorf("an unknown address should report ErrNoAccount, got %v", err)
	}
}

func TestLiveGoogleAccountLinksByEmailThenBySubject(t *testing.T) {
	ctx, pool, db := liveDB(t)
	email := uniqueEmail()
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM accounts WHERE email_normalised = $1`, auth.NormaliseEmail(email))
	})

	// Someone registers with a password first.
	hash, _ := auth.HashPassword("a-reasonable-passphrase")
	existing, err := db.CreateEmailAccount(ctx, email, hash, "Test Person")
	if err != nil {
		t.Fatal(err)
	}

	// Then signs in with Google using the same address: one person, one account.
	identity := auth.Identity{
		Subject: "google-sub-" + email, Email: email, EmailVerified: true, Name: "Test Person",
	}
	linked, created, err := db.UpsertGoogleAccount(ctx, identity)
	if err != nil {
		t.Fatalf("UpsertGoogleAccount: %v", err)
	}
	if created {
		t.Error("an existing address should be linked, not duplicated")
	}
	if linked != existing {
		t.Errorf("linked a different account: %s vs %s", linked, existing)
	}

	// Signing in again finds it by subject, even if the address changed.
	identity.Email = "changed_" + email
	again, created, err := db.UpsertGoogleAccount(ctx, identity)
	if err != nil {
		t.Fatalf("UpsertGoogleAccount (second): %v", err)
	}
	if created || again != existing {
		t.Errorf("the Google subject should find the same account: %s created=%v", again, created)
	}
}

func TestLiveGoogleAccountCreatesWhenUnknown(t *testing.T) {
	ctx, pool, db := liveDB(t)
	email := uniqueEmail()
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM accounts WHERE email_normalised = $1`, auth.NormaliseEmail(email))
	})

	id, created, err := db.UpsertGoogleAccount(ctx, auth.Identity{
		Subject: "brand-new-" + email, Email: email, EmailVerified: true, Name: "New Person",
	})
	if err != nil {
		t.Fatalf("UpsertGoogleAccount: %v", err)
	}
	if !created {
		t.Error("an unknown Google identity should create an account")
	}

	info, err := db.GetAccount(ctx, id)
	if err != nil {
		t.Fatalf("GetAccount: %v", err)
	}
	if info.Email != auth.NormaliseEmail(email) {
		t.Errorf("email: got %q", info.Email)
	}
	if info.DisplayName != "New Person" {
		t.Errorf("display name: got %q", info.DisplayName)
	}
	if len(info.AuthProviders) == 0 {
		t.Error("the account should record how it can sign in")
	}
}
