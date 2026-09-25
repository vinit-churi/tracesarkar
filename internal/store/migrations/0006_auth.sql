-- Email/password and Google identities.
--
-- ADR 0011 chose phone-only identity and rejected Aadhaar. This adds email and
-- Google as additional routes, not replacements: the reasoning and what it
-- costs are in ADR 0017. Phone verification remains the route for citizen
-- reporting, where a low-friction, widely-available identity matters most.
--
-- The password hash is bcrypt. The email is stored because authentication
-- needs it; it is personal data and lives under the same rules as everything
-- else the platform holds.

ALTER TABLE accounts ADD COLUMN IF NOT EXISTS email text;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS email_normalised text;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS password_hash bytea;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS google_sub text;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS display_name text;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS auth_providers text[] NOT NULL DEFAULT '{}';

-- One account per email, compared case-insensitively, and one per Google
-- subject. Partial indexes so the many accounts with neither stay unconstrained.
CREATE UNIQUE INDEX IF NOT EXISTS accounts_email_idx
  ON accounts (email_normalised) WHERE email_normalised IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS accounts_google_idx
  ON accounts (google_sub) WHERE google_sub IS NOT NULL;
