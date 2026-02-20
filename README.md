# Elevare Auction SaaS (MVP Backend)

Multi-tenant auction SaaS backend in Go using Chi + PostgreSQL + pgx + JWT + Argon2id + WebSockets.

## Run

```bash
go mod tidy
go run ./cmd/api
```

Environment variables:

- `APP_ADDR` (default `:8080`)
- `DATABASE_URL` (default `postgres://postgres:postgres@localhost:5432/elevare-auction?sslmode=disable`)
- `JWT_SECRET` (default `dev-secret-change-me`)

## Frontend (React + Vite)

```bash
cd frontend
npm install
npm run dev
```

Optional frontend env vars (`frontend/.env`):

- `VITE_API_BASE_URL=http://localhost:8080`
- `VITE_WS_BASE_URL=ws://localhost:8080`

## Run With Docker

Start DB, run migrations, and start API:

```bash
docker compose up --build
```

Services:

- API: `http://localhost:8080`
- Postgres: `localhost:5432`
- DB name: `elevare-auction`
- Swagger UI: `http://localhost:8080/swagger/index.html`

Stop:

```bash
docker compose down
```

Stop and remove DB volume:

```bash
docker compose down -v
```

## Migrations

Use `golang-migrate` with files in `migrations/`.

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

## Swagger

- UI: `GET /swagger/index.html`
- OpenAPI YAML: `GET /swagger/openapi.yaml`

## Tenant Resolution

Tenant is resolved from the request `Host`:

- subdomain: `tenant.yourauction.com`
- custom domain: exact match in `tenants.custom_domain`

## API Endpoints (MVP)

Public:

- `GET /public/config`
- `GET /auctions`
- `GET /auctions/{id}`
- `GET /auctions/{id}/ws`

Auth:

- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/verify-email`

Admin:

- `POST /admin/auctions`
- `PATCH /admin/auctions/{id}/start`
- `PATCH /admin/auctions/{id}/pause`
- `PATCH /admin/auctions/{id}/end`
- `GET /admin/bidders`
- `PATCH /admin/bidders/{id}/approve`

Bidding:

- `POST /auctions/{id}/bid`

## Notes

- All protected routes require Bearer JWT.
- All repo queries are tenant-scoped.
- Bid placement runs inside one DB transaction with `SELECT ... FOR UPDATE`.
- Scheduler loop runs every 5 seconds and auto starts/ends auctions.
