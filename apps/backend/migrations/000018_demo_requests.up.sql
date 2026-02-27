CREATE TABLE IF NOT EXISTS demo_requests (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  institution TEXT NOT NULL,
  email TEXT NOT NULL,
  message TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
