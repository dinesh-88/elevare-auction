# Elevare Auction

Elevare is a multi-tenant online auction platform. An auction house signs up, gets its own branded site on a subdomain, and runs timed auctions where bidders place live bids that update in real time over WebSockets.

**Who it's for:** small and mid-sized auction houses that want a hosted bidding site without building their own.

**Status:** personal side project and experiment. A public demo is deployed at [elevareauctions.com](https://elevareauctions.com), but it has no real customers or users: the production database holds only my own account, seeded demo users and a seeded demo auction house. All development uses data from the seed script and the built-in demo bid simulator.

## Features

- Self-serve tenant signup with platform-admin approval
- Auctions and lots with images, live bidding, and proxy (max) bids
- Real-time bid updates over WebSockets, plus a demo bid simulator
- Bidder registration and onboarding approval
- Buyer's premium/fee rules, tax calculation, and invoices (with CSV export)
- Staff roles and permissions, audit log
- Transactional email (outbid, lot won, invoice ready, verification) in EN/DE/FR, with an outbox, retries, and rate limits
- Local disk or S3/Cloudflare R2 file storage

## Tech Stack

- **Backend:** Go, Chi, PostgreSQL (pgx), JWT + Argon2id auth, WebSockets, background worker
- **Frontend:** Next.js 15 (App Router), React 19, TypeScript
- **Infra:** Docker Compose for local development, Render for deployment, Resend for email

## Structure

- `apps/backend`: Go API (Chi + PostgreSQL + pgx + JWT + Argon2id + WebSockets)
- `apps/web`: Next.js app (App Router, SSR-ready public + admin)
- `docker-compose.yml`: local Postgres + backend API + migrations + Next.js web

## Backend Local Run

```bash
cd apps/backend
go mod tidy
go run ./cmd/api
```

Environment variables:

- `APP_ADDR` (default `:8080`)
- `DATABASE_URL` (default `postgres://postgres:postgres@localhost:5432/elevare-auction?sslmode=disable`)
- `JWT_SECRET` (default `dev-secret-change-me`)
- `JWT_TTL` (default `30m`)
- `PLATFORM_ADMIN_TOKEN` (default `dev-platform-admin-token`)
- `SITE_PREVIEW_TOKEN` (default `dev-site-preview-token`)
- `ACCESS_TOKEN_COOKIE_NAME` (default `access_token`)
- `CSRF_COOKIE_NAME` (default `csrf_token`)
- `COOKIE_SECURE` (`true`/`false`, default `false`)
- `TENANT_APP_SCHEME` (default `https`)
- `TENANT_APP_BASE_DOMAIN` (default `elevareauctions.com`)
- `TENANT_APP_PORT` (optional; appended when building tenant URLs in non-standard environments)
- `STORAGE_PROVIDER` (`local` or `s3`, default `local`)
- `UPLOADS_DIR` (used when `STORAGE_PROVIDER=local`, default `/tmp/elevare-uploads`)
- `S3_BUCKET` (required when `STORAGE_PROVIDER=s3`)
- `S3_REGION` (default `us-east-1`)
- `S3_ENDPOINT` (optional, required for Cloudflare R2)
- `S3_ACCESS_KEY_ID` / `S3_SECRET_ACCESS_KEY`
- `S3_PUBLIC_BASE_URL` (recommended public/CDN base URL)
- `S3_USE_PATH_STYLE` (`true`/`false`, default `true`)
- `DEMO_TENANT_SLUG` (enables the built-in bid simulator for exactly one demo tenant when set)
- `BID_SIM_INTERVAL_SECONDS` (default `12`)
- `BID_SIM_IDLE_THRESHOLD_SECONDS` (default `20`)
- `BID_SIM_MAX_EVENTS_PER_MINUTE` (default `6`)
- `EMAIL_PROVIDER` (default `resend`)
- `RESEND_API_KEY` (required when `EMAIL_PROVIDER=resend` and email sending is enabled)
- `RESEND_FROM_EMAIL` (required when `EMAIL_PROVIDER=resend` and email sending is enabled)
- `INTERNAL_EMAIL_SECRET` (required to call `POST /internal/process-email-retries`)
- `EMAIL_RL_RECIPIENT_10M` (default `3`)
- `EMAIL_RL_RECIPIENT_24H` (default `20`)
- `EMAIL_RL_TENANT_1H` (default `200`)
- `EMAIL_RL_OUTBID_COOLDOWN_SECONDS` (default `60`)
- `EMAIL_RL_LOT_CLOSING_COOLDOWN_SECONDS` (default `600`)
- `EMAIL_RL_CRITICAL_RECIPIENT_10M` (default `10`)
- `EMAIL_RL_CRITICAL_RECIPIENT_24H` (default `100`)
- `EMAIL_RL_CRITICAL_TENANT_1H` (default `1000`)

Cloudflare R2 example:

```bash
STORAGE_PROVIDER=s3
S3_BUCKET=your-bucket
S3_REGION=auto
S3_ENDPOINT=https://<accountid>.r2.cloudflarestorage.com
S3_ACCESS_KEY_ID=...
S3_SECRET_ACCESS_KEY=...
S3_PUBLIC_BASE_URL=https://pub-<hash>.r2.dev
S3_USE_PATH_STYLE=true
```

AWS S3 example:

```bash
STORAGE_PROVIDER=s3
S3_BUCKET=your-bucket
S3_REGION=us-east-1
S3_ACCESS_KEY_ID=...
S3_SECRET_ACCESS_KEY=...
S3_PUBLIC_BASE_URL=https://your-bucket.s3.amazonaws.com
S3_USE_PATH_STYLE=false
```

## Web (Next.js) Local Run

```bash
cd apps/web
npm install
npm run dev
```

Web env (`apps/web/.env.local`):

- `NEXT_PUBLIC_API_BASE_URL=http://localhost:8080`
- `NEXT_PUBLIC_WS_BASE_URL=ws://localhost:8080`
- `API_BASE_URL=http://localhost:8080`

Web onboarding routes:

- Public tenant signup UI: `http://localhost:5173/signup`
- Platform review UI: `http://localhost:5173/platform/review`

## Docker Run (from repo root)

```bash
docker compose up --build
```

Services:

- API: `http://localhost:8080`
- Worker: notification delivery processor (`/app/worker`)
- Web: `http://localhost:5173`
- Swagger UI: `http://localhost:8080/swagger/index.html`
- Postgres: `localhost:5432`
- DB name: `elevare-auction`

Stop:

```bash
docker compose down
```

Reset DB volume:

```bash
docker compose down -v
```

## Migrations (manual)

```bash
migrate -path apps/backend/migrations -database "$DATABASE_URL" up
```

## Render Automated Migrations

This repo includes `render.yaml` plus an embedded migration binary (`/app/migrate`) in the backend image.

- API service: runs `preDeployCommand: /app/migrate` before `/app/api`
- Worker service: starts normally (`/app/worker`) and does not run migrations

To enable:

1. In Render, use **Blueprint** deploy from this repo (reads `render.yaml`), or copy the same pre-deploy command into your existing API/worker services.
2. Ensure both services have `DATABASE_URL` set.
3. Redeploy. Migrations will run automatically and treat `no change` as success.

Notes:
- `/app/migrate` sets `x-no-lock=true` by default to avoid advisory lock issues behind poolers/proxies.
- If needed, disable this behavior by setting `MIGRATE_NO_LOCK=false` on the API service.

## Seed Demo Data

From `apps/backend`:

```bash
DATABASE_URL="postgres://postgres:postgres@localhost:5432/elevare-auction?sslmode=disable" go run ./cmd/seed
```

Creates a demo tenant, admin/bidders, auctions/lots/bids, fee rules, and a sample invoice with image URLs.

## Tenant Onboarding Flow (Self-Serve)

- Public signup endpoint: `POST /signup`
- Platform review endpoints:
  - `GET /platform/admin/tenants?status=PENDING`
  - `PATCH /platform/admin/tenants/{id}/approve`
  - `PATCH /platform/admin/tenants/{id}/reject`
- Platform admin auth: `Authorization: Bearer <PLATFORM_ADMIN_TOKEN>`


![alt text](image.png)
