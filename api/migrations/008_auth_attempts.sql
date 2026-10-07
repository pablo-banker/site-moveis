CREATE TABLE auth_attempts (
 key_hash text PRIMARY KEY,
 attempts integer NOT NULL CHECK(attempts>0),
 resets_at timestamptz NOT NULL
);
CREATE INDEX auth_attempts_expiration ON auth_attempts(resets_at);
