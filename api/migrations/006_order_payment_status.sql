ALTER TABLE orders DROP CONSTRAINT orders_status_check;
UPDATE orders SET status='paid' WHERE status='paid_demo';
ALTER TABLE orders ADD CONSTRAINT orders_status_check CHECK(status IN ('awaiting_payment','cancelled','paid'));
ALTER TABLE simulated_payments RENAME TO order_payments;
ALTER TABLE order_payments ADD COLUMN provider text NOT NULL DEFAULT 'simulated';
ALTER TABLE order_payments ADD COLUMN simulated boolean NOT NULL DEFAULT true;
