DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'email_outbox_status') THEN
    CREATE TYPE email_outbox_status AS ENUM ('pending', 'sent', 'failed');
  END IF;
END $$;

CREATE TABLE email_outbox (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id uuid NULL REFERENCES users(id) ON DELETE SET NULL,
  to_email text NOT NULL,
  subject text NOT NULL,
  html_body text NOT NULL,
  text_body text NOT NULL,
  provider text NULL,
  status email_outbox_status NOT NULL DEFAULT 'pending',
  attempt_count integer NOT NULL DEFAULT 0,
  last_error text NULL,
  next_attempt_at timestamptz NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  sent_at timestamptz NULL
);

CREATE INDEX email_outbox_retry_idx
  ON email_outbox (status, next_attempt_at, attempt_count, created_at);
