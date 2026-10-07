ALTER TABLE orders DROP CONSTRAINT orders_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_status_check CHECK(status IN ('awaiting_payment','cancelled','paid_demo'));
CREATE TABLE simulated_payments (
 order_id uuid PRIMARY KEY REFERENCES orders(id),
 outcome text NOT NULL CHECK(outcome IN ('approved','declined')),
 updated_at timestamptz NOT NULL DEFAULT now()
);
