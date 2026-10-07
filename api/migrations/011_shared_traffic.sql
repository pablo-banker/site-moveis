CREATE TABLE rate_buckets (
 key_hash text PRIMARY KEY,
 attempts integer NOT NULL CHECK(attempts>0),
 resets_at timestamptz NOT NULL
);
CREATE INDEX rate_buckets_expiration ON rate_buckets(resets_at);
CREATE TABLE address_cache (
 cep text PRIMARY KEY CHECK(cep ~ '^[0-9]{8}$'),
 address jsonb NOT NULL,
 expires_at timestamptz NOT NULL
);
