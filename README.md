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

## Frontend Local Run

```bash
cd apps/frontend
npm install
npm run dev
```

Frontend env (`apps/frontend/.env`):

- `VITE_API_BASE_URL=http://localhost:8080`
- `VITE_WS_BASE_URL=ws://localhost:8080`

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
