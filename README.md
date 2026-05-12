# Real-Time Polling Application

A real-time polling app with live-updating results, built with **React**, **Go**, **PostgreSQL**, **Redis**, and **Docker**.

## Architecture

```
public_net          backend_net
  │                     │
  ├── proxy (nginx)     │
  │   └── :80           │
  │                     │
  ├── frontend          │
  │   (nginx)           │
  │                     │
  ├── backend (Go) ─────┤
  │                     ├── postgres
  │                     ├── redis
```

- **Nginx proxy** — single point of entry on port 80, routes `/api/*` to the Go backend and `/` to the frontend.
- **Frontend** — React + Vite + shadcn, served via its own Nginx container (no host ports).
- **Backend** — Go API with `database/sql` + `lib/pq` and `go-redis`. Votes increment Redis counters instantly and persist to PostgreSQL. Duplicate IPs per poll are rejected.
- **PostgreSQL** — persistent vote storage, completely isolated on `backend_net`.
- **Redis** — in-memory cache/backup for vote counts, also isolated on `backend_net`.

## Quick Start

```bash
cp .env.example .env
docker compose up --build
```

Open http://localhost.

## API Endpoints

| Method | Path                     | Description                        |
|--------|--------------------------|------------------------------------|
| GET    | `/api/polls`             | List all polls                     |
| POST   | `/api/polls`             | Create a poll                      |
| GET    | `/api/polls/{id}`        | Get poll details with options      |
| POST   | `/api/polls/{id}/vote`   | Submit a vote (`{option_id}`)      |
| GET    | `/api/polls/{id}/results`| Get live vote counts               |

## Stack

- **Frontend:** React 19, Vite, shadcn (Nova/Base UI), Wouter, Recharts
- **Backend:** Go 1.26, `net/http` (Go 1.22+ ServeMux), `lib/pq`, `go-redis/v9`
- **Infra:** PostgreSQL 17, Redis 7, Nginx, Docker Compose
