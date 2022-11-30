CREATE TABLE links (
  id bigserial primary key,
  name text NOT NULL,
  url text NOT NULL,
  hash text NOT NULL,
  favicon text,
  screenshot text,
  scrape_count int NOT NULL DEFAULT 0,
  last_change_at timestamptz NOT NULL default now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
);
