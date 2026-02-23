# Notifications MVP Runbook

## Prerequisites

- `RESEND_API_KEY` set in environment (or `.env` used by docker compose)
- `RESEND_FROM_EMAIL` set in environment
- DB migrated through `000011_notifications`

## Start Services

```bash
docker compose up -d db migrate api worker
```

## Verification Flow

1. Register bidder A and bidder B in the same tenant.
2. Place a bid as bidder A on a live lot.
3. Place a higher bid as bidder B on the same lot.
4. Verify bidder A receives:
   - `OUTBIDDED` row in `notifications`
   - `EMAIL` row in `notification_jobs` with `status='PENDING'` then `SENT`
5. End the lot from admin endpoint.
6. Verify winner receives:
   - `LOT_WON` in-app notification
   - `EMAIL` job processed by worker and marked `SENT`
7. Issue invoice for the lot.
8. Verify winner receives:
   - `INVOICE_ISSUED` in-app notification
   - `EMAIL` job processed by worker and marked `SENT`

## Useful SQL Checks

```sql
SELECT id, tenant_id, user_id, type, title, read_at, created_at
FROM notifications
ORDER BY created_at DESC;

SELECT id, tenant_id, user_id, channel, event_type, dedupe_key, status, attempts, next_run_at, sent_at
FROM notification_jobs
ORDER BY created_at DESC;

SELECT tenant_id, user_id, event_type, email_enabled, in_app_enabled
FROM notification_preferences
ORDER BY event_type;
```

## API Endpoints

- `GET /me/notifications?limit=50&cursor=<cursor>`
- `POST /me/notifications/{id}/read`
- `POST /me/notifications/read-all`
- `GET /me/notification-preferences`
- `PUT /me/notification-preferences`
