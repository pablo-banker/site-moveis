ALTER TABLE customer_sessions ADD COLUMN created_at timestamptz NOT NULL DEFAULT now();
CREATE INDEX customer_sessions_created ON customer_sessions(customer_id,created_at);
