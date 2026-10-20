# Roadmap

Planned work for Elevare Auction, in rough priority order. I'll build it with Claude Code as my AI coding assistant, so estimates are in hours of hands-on work with Claude Code: prompting, reviewing the generated code, testing by hand, and fixing what the AI gets wrong. Each task is broken into subtasks; the subtask hours add up to the task estimate.

**Total estimate: ~75 hours**

| Phase | Estimate |
| --- | --- |
| [1. Production readiness and security](#phase-1-production-readiness-and-security-215-h) | 21.5 h |
| [2. CI, tooling and tests](#phase-2-ci-tooling-and-tests-17-h) | 17 h |
| [3. Bidding and auction features](#phase-3-bidding-and-auction-features-12-h) | 12 h |
| [4. Payments and settlement](#phase-4-payments-and-settlement-11-h) | 11 h |
| [5. Admin and tenant experience](#phase-5-admin-and-tenant-experience-95-h) | 9.5 h |
| [6. Documentation](#phase-6-documentation-45-h) | 4.5 h |

---

## Phase 1: Production readiness and security (~21.5 h)

### 1.1 Refuse to start with default secrets in production (2 h)

- [ ] Add an `APP_ENV` setting (`development` / `production`) to the backend config (0.5 h)
- [ ] In production, fail at startup if `JWT_SECRET`, `PLATFORM_ADMIN_TOKEN` or `SITE_PREVIEW_TOKEN` is unset, a dev default, or too short (0.5 h)
- [ ] Apply the same check in the worker and migrate binaries (0.5 h)
- [ ] Unit tests for the config validation; document `APP_ENV` in the README and `render.yaml` (0.5 h)

### 1.2 Rate limiting on sensitive endpoints (3.5 h)

- [ ] Choose an approach: in-memory token bucket per instance, or Postgres-backed if the API runs on several instances (0.5 h)
- [ ] Rate-limit middleware keyed by client IP, honouring the proxy's forwarded-for header (1 h)
- [ ] Per-account limits for login and bid placement (0.5 h)
- [ ] Apply to login, signup, register, accept-invite and bid routes, with limits set from env vars (0.5 h)
- [ ] Return `429` with `Retry-After`; show a friendly message in the web app (0.5 h)
- [ ] Tests for allowed, limited and reset cases (0.5 h)

### 1.3 Password reset flow (3.5 h)

- [ ] Migration: `password_reset_tokens` table (hashed token, user, expiry, used_at) (0.5 h)
- [ ] Use case: request a reset (always returns success so emails can't be probed) and queue the email via the outbox (0.5 h)
- [ ] Use case: confirm the reset with token and new password, mark the token used, invalidate old sessions (0.5 h)
- [ ] HTTP endpoints, OpenAPI entries and rate limiting (0.5 h)
- [ ] Web pages: "Forgot password" form and "Set new password" page, linked from the login screen (1 h)
- [ ] Tests for expired, reused and valid tokens (0.5 h)

### 1.4 Account lockout after failed logins (2 h)

- [ ] Track failed login attempts and the last failure time per user (migration) (0.5 h)
- [ ] Exponential backoff, then a temporary lock after N failures; reset on success (0.5 h)
- [ ] Audit log entries for lockouts; admin can unlock a bidder from the bidder profile page (0.5 h)
- [ ] Tests (0.5 h)

### 1.5 Structured logging and health checks (2.5 h)

- [ ] Replace `log` calls with `log/slog` (JSON in production, text locally) (1 h)
- [ ] Request ID middleware; include request ID, tenant and user in every log line (0.5 h)
- [ ] `/healthz` (process up) and `/readyz` (database reachable) endpoints (0.5 h)
- [ ] Point Render health checks at `/readyz`; return the request ID in error responses (0.5 h)

### 1.6 Security headers (2 h)

- [ ] API: HSTS, `X-Content-Type-Options`, `Referrer-Policy`, `X-Frame-Options` middleware (0.5 h)
- [ ] Web: Content Security Policy in Next.js config, allowing the API, WebSocket and image storage origins (1 h)
- [ ] Check every page still works (images, sockets, Stripe redirect) and fix violations (0.5 h)

### 1.7 WebSocket security review (2 h)

- [ ] Audit how sockets authenticate and how a tenant is resolved on connect (0.5 h)
- [ ] Make sure a client can only subscribe to lots in its own tenant; reject the rest (0.5 h)
- [ ] Check the `Origin` header and cap message size and connections per user (0.5 h)
- [ ] Tests for cross-tenant subscription attempts (0.5 h)

### 1.8 Two-factor authentication for staff (4 h)

- [ ] Migration: TOTP secret (encrypted), enabled flag and recovery codes per user (0.5 h)
- [ ] Enrol: generate a secret and QR code, confirm with a first code, show recovery codes (1 h)
- [ ] Login: second step after the password for users with 2FA on (0.5 h)
- [ ] Tenant setting to require 2FA for all staff (0.5 h)
- [ ] Web screens for enrol, verify and disable (1 h)
- [ ] Tests (0.5 h)

---

## Phase 2: CI, tooling and tests (~17 h)

### 2.1 GitHub Actions pipeline (2.5 h)

- [ ] Backend job: `go build`, `go vet`, `go test ./...` with module caching (0.5 h)
- [ ] Web job: `npm ci`, `npm run lint`, `tsc --noEmit`, `npm run build` (0.5 h)
- [ ] Run on pull requests and pushes to `main`; add a status badge to the README (0.5 h)
- [ ] Docker build check for the backend image (0.5 h)
- [ ] Fix anything that fails the first run (0.5 h)

### 2.2 Linting (3 h)

- [ ] Add `.golangci.yml` (errcheck, staticcheck, gosec, revive) (0.5 h)
- [ ] Fix Go lint findings (1 h)
- [ ] Add an ESLint config for the web app (Next.js + TypeScript rules) (0.5 h)
- [ ] Fix web lint findings (0.5 h)
- [ ] Add both linters to CI (0.5 h)

### 2.3 Postgres integration tests (4 h)

- [ ] Test helper that starts a throwaway Postgres (Postgres service in CI, Docker locally) and runs migrations (1 h)
- [ ] Fixtures for a tenant, users, an auction and lots (0.5 h)
- [ ] Tests for bid placement and proxy bid resolution, including concurrent bids (1 h)
- [ ] Tests for invoice generation, fees and tax (0.5 h)
- [ ] Tests for tenant isolation in the repositories (0.5 h)
- [ ] Add the integration job to CI (0.5 h)

### 2.4 HTTP handler tests (3 h)

- [ ] Test server helper with fake use cases and auth cookies (0.5 h)
- [ ] Auth: login, logout, CSRF and cookie flags (0.5 h)
- [ ] Bidding: place bid, permission errors, closed lot, outbid (1 h)
- [ ] Invoices: list, detail, CSV export and access control (0.5 h)
- [ ] Error response format matches the OpenAPI spec (0.5 h)

### 2.5 Frontend unit tests (2.5 h)

- [ ] Set up Vitest, Testing Library and an API client mock (0.5 h)
- [ ] `BidInput`: increments, validation, submit and error states (0.5 h)
- [ ] Login and register forms (0.5 h)
- [ ] Admin create/edit lot forms (0.5 h)
- [ ] Add `npm test` to CI (0.5 h)

### 2.6 End-to-end smoke test (2 h)

- [ ] Set up Playwright against `docker compose` with seeded data (0.5 h)
- [ ] Scenario: tenant signup → platform approval → create auction and lot (0.5 h)
- [ ] Scenario: bidder registers → places bid → lot closes → invoice appears (0.5 h)
- [ ] Run nightly in CI and keep traces from failed runs (0.5 h)

---

## Phase 3: Bidding and auction features (~12 h)

### 3.1 Lot search and filters (3.5 h)

- [ ] Backend: query parameters for text, category, price range and closing-soon, with pagination (1 h)
- [ ] Database indexes, including full-text search on lot title and description (0.5 h)
- [ ] Web: search box and filter panel on the auction listing and auction detail pages (1 h)
- [ ] Keep filters in the URL so results can be shared (0.5 h)
- [ ] Tests (0.5 h)

### 3.2 Watchlist UI (2.5 h)

- [ ] Check the existing watchlist endpoints; add add/remove if missing (0.5 h)
- [ ] "Watch" toggle on lot cards and the lot detail page (0.5 h)
- [ ] Watchlist section on the bidder dashboard with live prices (1 h)
- [ ] Tests (0.5 h)

### 3.3 Staggered lot closing (3 h)

- [ ] Auction settings: first lot close time and gap between lots (0.5 h)
- [ ] Scheduler closes each lot at its own time, working with the existing anti-sniping extension (1 h)
- [ ] Broadcast per-lot countdown and close events over WebSockets (0.5 h)
- [ ] Web: per-lot countdown on lot cards and the lot page (0.5 h)
- [ ] Tests for close order and extensions (0.5 h)

### 3.4 "Closing soon" alerts for watched lots (3 h)

- [ ] Worker job finds watched lots closing within a set window (0.5 h)
- [ ] Create in-app notifications and queue emails through the outbox, using the existing cooldowns (1 h)
- [ ] Email templates in EN, DE and FR (0.5 h)
- [ ] Bidder notification preferences (on/off per type) (0.5 h)
- [ ] Tests (0.5 h)

---

## Phase 4: Payments and settlement (~11 h)

### 4.1 Finish Stripe Checkout for invoices, test mode only (4.5 h)

- [ ] Review the existing Stripe gateway and how it's wired into the API (0.5 h)
- [ ] "Pay now" button on the bidder invoices page that creates a Checkout session (1 h)
- [ ] Webhook endpoint with signature check; mark invoices paid on `checkout.session.completed` (1 h)
- [ ] Idempotency: store processed event IDs and ignore duplicates (0.5 h)
- [ ] Success and cancel pages; show payment status on the invoice (0.5 h)
- [ ] Tests with recorded webhook payloads, and a local guide using the Stripe CLI (1 h)

### 4.2 Seller (consignor) statements (4 h)

- [ ] Data model: consignors, and linking lots to a consignor (migration) (1 h)
- [ ] Commission rules per consignor or per auction (0.5 h)
- [ ] Statement calculation: hammer total, commission, fees, payout due (1 h)
- [ ] Admin screen to view a statement and export it as CSV (1 h)
- [ ] Tests for the calculations (0.5 h)

### 4.3 Overdue invoice reminders (2.5 h)

- [ ] Tenant setting for the reminder schedule (for example 3, 7 and 14 days) (0.5 h)
- [ ] Worker job that finds unpaid invoices past their due date and queues reminders (1 h)
- [ ] Email templates in EN, DE and FR (0.5 h)
- [ ] Record reminders sent in the audit log; tests (0.5 h)

---

## Phase 5: Admin and tenant experience (~9.5 h)

### 5.1 Bulk lot import from CSV (4 h)

- [ ] Define the CSV format and publish a downloadable template (0.5 h)
- [ ] Backend: parse and validate rows, report errors per row, import in one transaction (1 h)
- [ ] Match images by filename from a bulk image upload (1 h)
- [ ] Admin screen: upload, preview, fix errors, confirm (1 h)
- [ ] Tests (0.5 h)

### 5.2 Tenant analytics dashboard (3 h)

- [ ] Backend queries: registrations over time, bids per lot, sell-through rate, revenue per auction (1 h)
- [ ] Analytics endpoint with a date range filter (0.5 h)
- [ ] Charts and summary numbers on the admin dashboard (1 h)
- [ ] Tests (0.5 h)

### 5.3 DE and FR translations in the web app (2.5 h)

- [ ] Find hard-coded strings left in the screens and move them into translation files (1 h)
- [ ] Translate and review DE and FR (1 h)
- [ ] Language switcher check and fallbacks for missing keys (0.5 h)

---

## Phase 6: Documentation (~4.5 h)

### 6.1 Architecture overview (1.5 h)

- [ ] `docs/architecture.md`: backend layers (domain, use case, adapters, transport) (0.5 h)
- [ ] How multi-tenancy works: tenant resolution, isolation, subdomains (0.5 h)
- [ ] Diagram of the bid flow: HTTP → use case → Postgres → WebSocket broadcast → email outbox (0.5 h)

### 6.2 OpenAPI spec (1.5 h)

- [ ] Compare `openapi.yaml` with the router and add missing endpoints (0.5 h)
- [ ] Request and response examples for the main flows (0.5 h)
- [ ] CI check that every route in the router appears in the spec (0.5 h)

### 6.3 Contributor guide (1.5 h)

- [ ] `CONTRIBUTING.md`: local setup, running tests and linters (0.5 h)
- [ ] Migration conventions and how to add an email template (0.5 h)
- [ ] Pull request checklist (0.5 h)

---

## Notes

- This is a personal experiment with no real customers. All data is seeded or simulated.
- Stripe work stays in test mode; no real payments are processed.
