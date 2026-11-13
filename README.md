# ProHealth API

Multi-tenant clinic booking backend with an **AI booking agent**. Patients chat in natural language; the agent finds open slots, proposes one, and books it **only after the patient confirms**.

Go · Gin · GORM · PostgreSQL 16 · Redis 7 · OpenAI (GPT-4o) · Docker

---

## Highlights

**AI booking agent (tool-calling)**
- Three tools exposed to the model: `search_available_slots` → `propose_slot` → `create_appointment`.
- `create_appointment` is only called after the patient explicitly confirms the proposed slot — the model can suggest, but never books on its own.
- Provider-agnostic `LLM` interface (`internal/llm`), with the OpenAI provider wrapped in a **circuit breaker** (`sony/gobreaker`) so an upstream outage fails fast instead of piling up requests.
- Per-session conversation history; the system prompt is built from the hospital's real departments and the known patient.

**Multi-tenancy**
- Every tenant-scoped request carries `X-Hospital-ID`; `TenantMiddleware` checks the caller's membership role in that hospital before the handler runs.
- Hospital invitations, system roles vs. hospital roles, user/hospital status, session table, audit log.

**Reliability**
- **Idempotency middleware** (Redis): hashes the request, tracks `processing` / `completed` state, and replays the stored response on retry — a double-clicked "Book" never creates two appointments.
- **Token-bucket rate limiting** as an atomic Redis Lua script (auth: 10 req/min, API: 200 req/min).
- **Distributed lock** (Redis `SET NX` + Lua compare-and-delete unlock).
- **Transactional outbox → Redis Streams → notification consumer**: side effects (appointment e-mails) are written in the same transaction as the booking, published by a worker, and consumed by a separate process with bounded retries.
- Request IDs, structured logging, per-route timeouts.

---

## Architecture

```
            ┌──────────────────────────── cmd/server ────────────────────────────┐
 client ──▶ │ request-id → timeout → rate-limit → auth → tenant → role → handler │
            │                                   │                                │
            │        services ──▶ PostgreSQL ◀── outbox_jobs (same tx)           │
            │           │                            │                           │
            │      AgentService ──▶ LLM (OpenAI + circuit breaker)               │
            │                                 OutboxWorker ──▶ Redis Stream      │
            └────────────────────────────────────────────────────┬───────────────┘
                                                                 ▼
                                          cmd/notification: consumer group ──▶ SMTP
```

## Project layout

```
cmd/
  server/         HTTP API + outbox worker
  notification/   Redis Streams consumer (e-mail)
  seed/           demo data
internal/
  handler/ service/ model/ dto/ response/
  middleware/     auth, tenant, role, idempotency, rate limit, timeout, request id, logging, cors
  llm/            LLM interface + OpenAI provider
  outbox/         publisher + worker
  notification/   consumer + dispatcher
  ratelimit/      token bucket (Lua)
  lock/           distributed lock (Lua)
  audit/ cache/ security/ config/ container/
db/migrations/    golang-migrate SQL migrations
routes/           route registration per module
```

## API modules

All routes live under `/api/v1`.

| Module | Examples |
|---|---|
| Auth | register, login, change password |
| Hospitals & invitations | create hospital, invite members, accept invitation |
| Departments & schedules | departments, schedule periods, shifts |
| Appointments | create (idempotent), list (admin), my appointments, update |
| Chat | `POST /chat` — AI booking agent |
| Admin | users, hospitals, audit logs |
| Health | `GET /health` |

---

## Run locally

Requirements: Go 1.25+, Docker, [golang-migrate](https://github.com/golang-migrate/migrate), optionally [air](https://github.com/air-verse/air) for hot reload.

```bash
cp .env.example .env          # fill in the values
docker compose up -d db redis
make migrate-up
go run ./cmd/seed             # optional demo data
go run ./cmd/server           # API
go run ./cmd/notification     # e-mail consumer (separate terminal)
```

Or run everything in containers:

```bash
docker compose up --build
```

### Environment variables

| Variable | Purpose |
|---|---|
| `PORT` | HTTP port |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DATABASE_URL` | PostgreSQL |
| `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD` | Redis |
| `JWT_SECRET` | Token signing |
| `OPENAI_API_KEY`, `OPENAI_MODEL` | AI booking agent |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `MAIL_FROM` | Outgoing e-mail |
| `CONSUMER_NAME` | Redis Streams consumer name |
| `CORS_ORIGINS`, `FRONTEND_URL` | Web client |

---

## Related

- **SmartClinic** — the v2 of this project: a full hospital management system (patient records, inpatient/outpatient encounters, lab & imaging, pharmacy, inventory, billing, insurance claims).
- Web client: React (separate repository).
