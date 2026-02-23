DROP TABLE IF EXISTS notification_preferences;
DROP INDEX IF EXISTS idx_notification_jobs_status_next_run;
DROP INDEX IF EXISTS uq_notification_jobs_dedupe_key;
DROP TABLE IF EXISTS notification_jobs;
DROP INDEX IF EXISTS idx_notifications_tenant_user_created;
DROP TABLE IF EXISTS notifications;
