# Elevare Auction Monorepo

## Structure

- `apps/backend`: Go API (Chi + PostgreSQL + pgx + JWT + Argon2id + WebSockets)
- `apps/frontend`: React + Vite + TypeScript app (public + admin)
- `docker-compose.yml`: local Postgres + backend API + migrations

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

## Frontend Local Run

```bash
cd apps/frontend
npm install
npm run dev
```

Frontend env (`apps/frontend/.env`):

- `VITE_API_BASE_URL=http://localhost:8080`
- `VITE_WS_BASE_URL=ws://localhost:8080`

Frontend onboarding routes:

- Public tenant signup UI: `http://localhost:5173/signup`
- Platform review UI: `http://localhost:5173/platform/review`

## Docker Run (from repo root)

```bash
docker compose up --build
```

Services:

- API: `http://localhost:8080`
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

## Tenant Onboarding Flow (Self-Serve)

- Public signup endpoint: `POST /signup`
- Platform review endpoints:
  - `GET /platform/admin/tenants?status=PENDING`
  - `PATCH /platform/admin/tenants/{id}/approve`
  - `PATCH /platform/admin/tenants/{id}/reject`
- Platform admin auth: `Authorization: Bearer <PLATFORM_ADMIN_TOKEN>`


![alt text](image.png)
