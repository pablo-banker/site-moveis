CREATE TABLE categories (
 id text PRIMARY KEY,
 name text NOT NULL UNIQUE,
 position integer NOT NULL DEFAULT 0
);
CREATE TABLE rooms (
 id text PRIMARY KEY,
 name text NOT NULL UNIQUE,
 position integer NOT NULL DEFAULT 0
);
CREATE TABLE products (
 id text PRIMARY KEY,
 name text NOT NULL,
 category_id text NOT NULL REFERENCES categories(id),
 room_id text NOT NULL REFERENCES rooms(id),
 price_cents bigint NOT NULL CHECK(price_cents >= 0),
 image text NOT NULL,
 dimensions text NOT NULL,
 material text NOT NULL,
 label text NOT NULL DEFAULT '',
 finishes text[] NOT NULL CHECK(cardinality(finishes) > 0),
 position integer NOT NULL DEFAULT 0,
 active boolean NOT NULL DEFAULT true,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX products_category ON products(category_id);
CREATE INDEX products_room ON products(room_id);
CREATE INDEX products_price ON products(price_cents,id) WHERE active;
