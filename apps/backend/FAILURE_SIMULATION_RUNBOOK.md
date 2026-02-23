# Failure Simulation Runbook (Break It On Purpose)

This is a concrete execution guide for the current Elevare backend/web stack.

## Prerequisites

- Start services:
  - `docker compose up -d db migrate api worker web`
- Seed demo data:
  - `cd apps/backend && go run ./cmd/seed`
- Demo tenant host:
  - `demo.localhost:8080`

## Test Identities

- Admin: `admin@demo.elevare.test / admin123`
- Bidder A: `bidder@demo.elevare.test / bidder123`
- Bidder B: `bidder2@demo.elevare.test / bidder123`

## Seeded Scenario IDs

- Auction (LIVE): `61111111-1111-1111-1111-111111111111`
- Proxy scenario lot: `a1111111-1111-1111-1111-111111111111`
- Soft-close scenario lot: `b1111111-1111-1111-1111-111111111111`
- Tenant: `11111111-1111-1111-1111-111111111111`

## Helper Setup

```bash
API="http://localhost:8080"
HOST="demo.localhost:8080"
COOKIE_A="/tmp/elevare_a.cookies"
COOKIE_B="/tmp/elevare_b.cookies"
COOKIE_ADMIN="/tmp/elevare_admin.cookies"
```

Login helper:

```bash
login() {
  local email="$1"
  local pass="$2"
  local jar="$3"
  curl -s -c "$jar" -H "Host: $HOST" -H "Content-Type: application/json" \
    -d "{\"email\":\"$email\",\"password\":\"$pass\"}" \
    "$API/auth/login"
}
```

Extract CSRF token from cookie jar:

```bash
csrf_from_cookiejar() {
  awk '$6=="csrf_token"{print $7}' "$1" | tail -n1
}
```

## A) Bidding & Concurrency

### A1 Concurrent bids (same second)

```bash
login "bidder@demo.elevare.test" "bidder123" "$COOKIE_A" >/dev/null
login "bidder2@demo.elevare.test" "bidder123" "$COOKIE_B" >/dev/null
CSRF_A=$(csrf_from_cookiejar "$COOKIE_A")
CSRF_B=$(csrf_from_cookiejar "$COOKIE_B")

# Run in parallel
(
  curl -s -b "$COOKIE_A" -H "Host: $HOST" -H "X-CSRF-Token: $CSRF_A" -H "Content-Type: application/json" \
    -d '{"bid_amount_minor":9300}' "$API/lots/$LOT_ID/bid"
) &
(
  curl -s -b "$COOKIE_B" -H "Host: $HOST" -H "X-CSRF-Token: $CSRF_B" -H "Content-Type: application/json" \
    -d '{"bid_amount_minor":9300}' "$API/lots/$LOT_ID/bid"
) &
wait
```

Verify:

```sql
SELECT id, user_id, amount, created_at
FROM lot_bids
WHERE tenant_id='11111111-1111-1111-1111-111111111111' AND lot_id='<LOT_ID>'
ORDER BY created_at DESC;
```

### A2 Proxy vs manual race

- Use `LOT_ID=a1111111-1111-1111-1111-111111111111`.
- A places max, B places manual quickly after.

```bash
curl -s -b "$COOKIE_A" -H "Host: $HOST" -H "X-CSRF-Token: $CSRF_A" -H "Content-Type: application/json" \
  -d '{"max_bid_minor":10000}' "$API/lots/a1111111-1111-1111-1111-111111111111/bid"

curl -s -b "$COOKIE_B" -H "Host: $HOST" -H "X-CSRF-Token: $CSRF_B" -H "Content-Type: application/json" \
  -d '{"bid_amount_minor":8500}' "$API/lots/a1111111-1111-1111-1111-111111111111/bid"
```

### A3 Proxy vs proxy competition

```bash
curl -s -b "$COOKIE_A" -H "Host: $HOST" -H "X-CSRF-Token: $CSRF_A" -H "Content-Type: application/json" \
  -d '{"max_bid_minor":10000}' "$API/lots/$LOT_ID/bid"

curl -s -b "$COOKIE_B" -H "Host: $HOST" -H "X-CSRF-Token: $CSRF_B" -H "Content-Type: application/json" \
  -d '{"max_bid_minor":9000}' "$API/lots/$LOT_ID/bid"
```

Public leakage check:

```bash
curl -s -H "Host: $HOST" "$API/lots/$LOT_ID/bids"
```

- Confirm response has no `max_bid_minor`, no `bid_type`.

### A4 Boundary bid at end time

- Place bid when countdown is near zero in UI or after forcing lot `ended_at`.
- Expected: reject when `now() >= end_time`.

## B) Soft Close

### B1 Extension triggers

- `LOT_ID=b1111111-1111-1111-1111-111111111111` is seeded near end.

```bash
curl -s -b "$COOKIE_A" -H "Host: $HOST" -H "X-CSRF-Token: $CSRF_A" -H "Content-Type: application/json" \
  -d '{"bid_amount_minor":9400}' "$API/lots/b1111111-1111-1111-1111-111111111111/bid"
```

Verify:

```sql
SELECT id, ended_at
FROM lots
WHERE id='b1111111-1111-1111-1111-111111111111';
```

### B2 Multiple extensions

- Place multiple bids within window and ensure `ended_at` moves forward monotonically.

### B3 Soft close disabled

```bash
login "admin@demo.elevare.test" "admin123" "$COOKIE_ADMIN" >/dev/null
CSRF_ADMIN=$(csrf_from_cookiejar "$COOKIE_ADMIN")

curl -s -X PATCH -b "$COOKIE_ADMIN" \
  -H "Host: $HOST" -H "X-CSRF-Token: $CSRF_ADMIN" -H "Content-Type: application/json" \
  -d '{"auto_extend_enabled":false}' \
  "$API/admin/auctions/61111111-1111-1111-1111-111111111111"
```

## C) Worker & Notifications

### C1 Worker down, then recover

```bash
docker compose stop worker
# trigger outbid/lot-won/invoice-issued events
docker compose start worker
```

Verify jobs:

```sql
SELECT id, event_type, status, attempts, next_run_at, sent_at
FROM notification_jobs
WHERE tenant_id='11111111-1111-1111-1111-111111111111'
ORDER BY created_at DESC;
```

### C2 Email provider failure

- Set wrong `RESEND_API_KEY`, restart worker, trigger event.
- Verify retries and final `FAILED` after max attempts.

### C3 Duplicate emit dedupe

- Trigger same event twice with same dedupe reference.

```sql
SELECT dedupe_key, count(*)
FROM notification_jobs
GROUP BY dedupe_key
HAVING count(*) > 1;
```

## D) DB/Transaction Faults

### D1 DB outage mid-bid

```bash
docker compose stop db
# attempt bid (expect 500/failure)
docker compose start db
```

Verify no partial write:

```sql
SELECT current_price FROM lots WHERE id='<LOT_ID>';
SELECT count(*) FROM lot_bids WHERE tenant_id='11111111-1111-1111-1111-111111111111' AND lot_id='<LOT_ID>';
```

### D2 Deadlock smoke

- Run rapid parallel bids from both bidders in a loop; check logs for deadlocks, confirm no inconsistent final state.

## E) Auth & Tenant Isolation

### E1 Cross-tenant access

- Use tenant A session with tenant B host; expect `401/403/404`.

### E2 Suspended bidder tries to bid

- Suspend bidder, then attempt bid.
- Expected target: reject with forbidden and no bid row.

### E3 Expired session

- Remove/expire cookie and call `mark-paid`; expect `401`.

## F) Settlement / Invoice Integrity

### F1 Mark paid twice

```bash
curl -s -X POST -b "$COOKIE_ADMIN" -H "Host: $HOST" -H "X-CSRF-Token: $CSRF_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"payment_reference":"TEST-REF-1"}' \
  "$API/admin/invoices/<INVOICE_ID>/mark-paid"
```

- Repeat and verify safe behavior.

### F2 Cancel paid invoice

```bash
curl -s -X POST -b "$COOKIE_ADMIN" -H "Host: $HOST" -H "X-CSRF-Token: $CSRF_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"reason":"test"}' \
  "$API/admin/invoices/<INVOICE_ID>/cancel"
```

- Expect rejection if paid.

### F3 CSV export with cancelled

```bash
curl -i -b "$COOKIE_ADMIN" -H "Host: $HOST" -H "X-CSRF-Token: $CSRF_ADMIN" \
  "$API/admin/auctions/61111111-1111-1111-1111-111111111111/export-settlement-csv"
```

## Audit Evidence Queries

```sql
SELECT action, entity_type, entity_id, metadata, created_at
FROM audit_logs
WHERE tenant_id='11111111-1111-1111-1111-111111111111'
ORDER BY created_at DESC
LIMIT 200;
```

Filter critical actions:

```sql
SELECT action, entity_type, entity_id, created_at
FROM audit_logs
WHERE tenant_id='11111111-1111-1111-1111-111111111111'
  AND action IN (
    'BID_PLACED',
    'LOT_EXTENDED',
    'LOT_ENDED',
    'INVOICE_ISSUED',
    'INVOICE_MARKED_PAID',
    'INVOICE_CANCELLED',
    'SETTLEMENT_CSV_EXPORTED',
    'NOTIFICATION_SENT',
    'NOTIFICATION_FAILED',
    'SITE_PUBLISHED'
  )
ORDER BY created_at DESC;
```

## Results Template

- Scenario:
- Pass/Fail:
- Request IDs:
- Logs excerpt:
- Audit evidence:
- Data integrity check:
