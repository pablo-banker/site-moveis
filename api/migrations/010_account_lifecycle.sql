ALTER TABLE customers ADD COLUMN email_verified_at timestamptz;
CREATE TABLE account_challenges (
 token_hash text PRIMARY KEY,
 customer_id uuid NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
 purpose text NOT NULL CHECK(purpose IN ('verify','reset')),
 expires_at timestamptz NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX account_challenges_customer ON account_challenges(customer_id,purpose);
CREATE TABLE email_outbox (
 id uuid PRIMARY KEY,
 recipient text NOT NULL,
 subject text NOT NULL,
 body text NOT NULL,
 attempts integer NOT NULL DEFAULT 0,
 available_at timestamptz NOT NULL DEFAULT now(),
 sent_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX email_outbox_pending ON email_outbox(available_at) WHERE sent_at IS NULL;
CREATE TABLE security_events (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 customer_id uuid REFERENCES customers(id) ON DELETE SET NULL,
 event text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
