CREATE TABLE contact_messages (
 id uuid PRIMARY KEY,
 name text NOT NULL,
 email text NOT NULL,
 subject text NOT NULL,
 body text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
