CREATE TABLE shipping_quotes (
 id uuid PRIMARY KEY,
 customer_id uuid NOT NULL REFERENCES customers(id),
 fingerprint text NOT NULL,
 quote jsonb NOT NULL,
 expires_at timestamptz NOT NULL
);
CREATE INDEX shipping_quotes_expiration ON shipping_quotes(expires_at);
ALTER TABLE orders ADD COLUMN shipping jsonb;
