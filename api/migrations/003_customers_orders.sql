CREATE TABLE customers (
 id uuid PRIMARY KEY,
 name text NOT NULL,
 email text NOT NULL UNIQUE,
 password_hash text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE customer_sessions (
 token_hash text PRIMARY KEY,
 customer_id uuid NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
 expires_at timestamptz NOT NULL
);
CREATE INDEX customer_sessions_customer ON customer_sessions(customer_id);
CREATE TABLE product_variants (
 product_id text NOT NULL REFERENCES products(id),
 finish text NOT NULL,
 stock integer NOT NULL CHECK(stock>=0),
 PRIMARY KEY(product_id,finish)
);
-- Demonstration stock, not commercial availability.
INSERT INTO product_variants(product_id,finish,stock)
SELECT p.id,f,5 FROM products p CROSS JOIN LATERAL unnest(p.finishes) f;
CREATE TABLE orders (
 id uuid PRIMARY KEY,
 customer_id uuid NOT NULL REFERENCES customers(id),
 idempotency_key uuid NOT NULL,
 request_hash text NOT NULL,
 status text NOT NULL CHECK(status IN ('awaiting_payment','cancelled')),
 subtotal_cents bigint NOT NULL CHECK(subtotal_cents>=0),
 shipping_cents bigint NOT NULL CHECK(shipping_cents>=0),
 total_cents bigint NOT NULL CHECK(total_cents=subtotal_cents+shipping_cents),
 address jsonb NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(customer_id,idempotency_key)
);
CREATE TABLE order_items (
 order_id uuid NOT NULL REFERENCES orders(id),
 product_id text NOT NULL REFERENCES products(id),
 finish text NOT NULL,
 name text NOT NULL,
 quantity integer NOT NULL CHECK(quantity BETWEEN 1 AND 10),
 unit_price_cents bigint NOT NULL CHECK(unit_price_cents>=0),
 PRIMARY KEY(order_id,product_id,finish)
);
CREATE INDEX orders_customer ON orders(customer_id,created_at DESC);
