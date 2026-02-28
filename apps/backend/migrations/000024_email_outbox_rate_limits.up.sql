ALTER TABLE email_outbox
  ADD COLUMN IF NOT EXISTS type text NULL,
  ADD COLUMN IF NOT EXISTS entity_type text NULL,
  ADD COLUMN IF NOT EXISTS entity_id uuid NULL;

CREATE INDEX IF NOT EXISTS email_outbox_recipient_created_idx
  ON email_outbox (tenant_id, to_email, created_at);

CREATE INDEX IF NOT EXISTS email_outbox_tenant_created_idx
  ON email_outbox (tenant_id, created_at);

CREATE INDEX IF NOT EXISTS email_outbox_scope_cooldown_idx
  ON email_outbox (tenant_id, type, entity_type, entity_id, user_id, created_at);
