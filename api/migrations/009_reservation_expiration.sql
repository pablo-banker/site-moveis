ALTER TABLE orders ADD COLUMN reservation_expires_at timestamptz NOT NULL DEFAULT now()+interval '30 minutes';
UPDATE orders SET reservation_expires_at=created_at+interval '30 minutes';
CREATE INDEX orders_pending_expiration ON orders(reservation_expires_at) WHERE status='awaiting_payment';
