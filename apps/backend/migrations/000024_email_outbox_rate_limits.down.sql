DROP INDEX IF EXISTS email_outbox_scope_cooldown_idx;
DROP INDEX IF EXISTS email_outbox_tenant_created_idx;
DROP INDEX IF EXISTS email_outbox_recipient_created_idx;

ALTER TABLE email_outbox
  DROP COLUMN IF EXISTS entity_id,
  DROP COLUMN IF EXISTS entity_type,
  DROP COLUMN IF EXISTS type;
