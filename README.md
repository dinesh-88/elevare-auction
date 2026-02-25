# Elevare Auction Monorepo

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
- `PLATFORM_ADMIN_TOKEN` (default `dev-platform-admin-token`)
- `STORAGE_PROVIDER` (`local` or `s3`, default `local`)
- `UPLOADS_DIR` (used when `STORAGE_PROVIDER=local`, default `/tmp/elevare-uploads`)
- `S3_BUCKET` (required when `STORAGE_PROVIDER=s3`)
- `S3_REGION` (default `us-east-1`)
- `S3_ENDPOINT` (optional, required for Cloudflare R2)
- `S3_ACCESS_KEY_ID` / `S3_SECRET_ACCESS_KEY`
- `S3_PUBLIC_BASE_URL` (recommended public/CDN base URL)
- `S3_USE_PATH_STYLE` (`true`/`false`, default `true`)
- `RESEND_API_KEY` (required for notification email worker)
- `RESEND_FROM_EMAIL` (required for notification email worker)

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
