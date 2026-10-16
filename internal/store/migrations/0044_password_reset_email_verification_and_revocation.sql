-- 0044: the three things system 5 was missing.
--
-- ADR 0017 shipped email/password and Google sign-in and listed what it did
-- not do. Three of those are here: a password that can be reset without a
-- human in the loop, an address that can be proved to reach the person
-- holding it, and a session that can be ended from the server rather than
-- waited out.
--
-- **Revocation, and why a number rather than a list.** A session token is a
-- signed JWT, which is what makes it cheap: nothing is read from the database
-- to authenticate a request. The cost is that a signed token cannot be
-- withdrawn. The usual answers are a deny list of revoked tokens, which grows
-- without bound and has to be consulted anyway, or server-side sessions,
-- which gives up the property that made the JWT worth having. The cheaper
-- answer is a version: the account holds a number, every token carries the
-- number it was minted with, and bumping the account's number invalidates
-- every token in existence at once. One integer, one comparison, no growth,
-- and "sign out everywhere" and "change the password" are the same operation
-- seen from two sides.
--
-- The comparison is made on every authenticated request, so it does read the
-- account — but it reads one row by primary key, which the API already does
-- for most authenticated work, and the alternative was a list that grows
-- forever.

ALTER TABLE accounts ADD COLUMN IF NOT EXISTS token_version bigint NOT NULL DEFAULT 0;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS email_verified_at timestamptz;

COMMENT ON COLUMN accounts.token_version IS
  'Bumped to invalidate every session token issued before now. A token '
  'carries the version it was minted with; a mismatch is an expired session.';
COMMENT ON COLUMN accounts.email_verified_at IS
  'When the address was proved to reach the person holding this account. Null '
  'means unproved, which is not the same as absent.';

-- One-time credentials: password reset and email verification.
--
-- Only the hash is stored. A reset link that leaked from a backup would
-- otherwise be a password reset for everybody who had one pending, and a
-- backup is exactly the thing that leaks. The hash is SHA-256 rather than
-- bcrypt on purpose: the token is 256 random bits, so there is nothing to
-- slow an attacker down from guessing, and a work factor would buy latency
-- and no safety (internal/auth/onetime.go).
--
-- `used_at` rather than a delete, so that presenting a spent link gives the
-- same answer as presenting a forged one and the row survives for an audit.
CREATE TABLE IF NOT EXISTS auth_tokens (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id  uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    purpose     text NOT NULL,
    token_hash  bytea NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz NOT NULL,
    used_at     timestamptz,

    CONSTRAINT auth_tokens_purpose_known
      CHECK (purpose IN ('password_reset', 'email_verification')),
    CONSTRAINT auth_tokens_expires_after_creation
      CHECK (expires_at > created_at)
);

-- The lookup is by hash and purpose, and nothing else: a credential minted to
-- verify an address must not be presentable to reset a password. Unique on
-- the hash so a collision is a constraint violation rather than two accounts
-- sharing a credential.
CREATE UNIQUE INDEX IF NOT EXISTS auth_tokens_hash_idx
  ON auth_tokens (token_hash);
CREATE INDEX IF NOT EXISTS auth_tokens_account_idx
  ON auth_tokens (account_id, purpose, created_at DESC);

-- No stored IP address and no user agent. Who asked for a reset is not
-- something this platform needs to keep, and a request-rate limit already
-- lives in internal/ratelimit where it does not become a record of who tried
-- to sign in from where.

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
     WHERE table_name = 'accounts' AND column_name = 'token_version') THEN
    RAISE EXCEPTION 'accounts.token_version was not created';
  END IF;
END $$;
