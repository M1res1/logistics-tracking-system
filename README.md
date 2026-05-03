# logistics-tracking-system
logistics-tracking-system web app

# Food App — Infrastructure

Get a full local environment running in under 5 minutes.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) ≥ 24
- [Docker Compose](https://docs.docker.com/compose/) ≥ 2.20
- [Go](https://go.dev/dl/) ≥ 1.22 (for running tests locally)
- Make

---

## Quick Start

```bash
# 1. Clone
git clone https://github.com/your-org/food-app.git
cd food-app

# 2. Configure environment
cp .env.example .env
# Edit .env if needed — defaults work out of the box for local dev

# 3. Boot infrastructure (postgres, redis, kafka)
make up

# 4. Run migrations
make migrate

# 5. Start everything (app services + nginx + observability)
docker compose up -d
```

App is now available at **http://localhost**.

---

## Services & Ports

| Service       | Port  | Path prefix        |
|---------------|-------|--------------------|
| Auth          | 8080  | `/api/v1/auth/*`   |
| Orders        | 8081  | `/orders/*`        |
| Deliveries    | 8082  | `/deliveries/*`    |
| Restaurants   | 8083  | `/restaurants/*`   |
| Payments      | 8084  | `/payments/*`      |
| Nginx (entry) | 80    | all of the above   |
| Prometheus    | 9090  | —                  |
| Grafana       | 3000  | — (admin/admin)    |

---

## Makefile Commands

| Command        | Description                              |
|----------------|------------------------------------------|
| `make up`      | Start postgres, redis, kafka             |
| `make down`    | Stop and remove all containers           |
| `make migrate` | Run SQL migrations in `./migrations/`    |
| `make test`    | Run all Go tests                         |
| `make logs`    | Tail all logs (`s=auth` to filter)       |
| `make ps`      | Show running containers                  |

---

## Project Structure

```
.
├── docker-compose.yml       # All services
├── Dockerfile               # Multi-stage Go build (shared template)
├── .env.example             # Environment variable reference
├── Makefile                 # Dev shortcuts
├── nginx/
│   └── nginx.conf           # Path-based routing
├── prometheus/
│   └── prometheus.yml       # Scrape config
├── grafana/
│   └── provisioning/        # Auto-loaded datasource + dashboard
├── migrations/              # SQL files, applied in sort order
└── services/
    ├── auth/cmd/main.go
    ├── orders/cmd/main.go
    ├── deliveries/cmd/main.go
    ├── restaurants/cmd/main.go
    └── payments/cmd/main.go
```

---

## Adding a Migration

Drop a `.sql` file into `./migrations/` with a numeric prefix and run `make migrate`:

```
migrations/
  001_create_users.sql
  002_create_orders.sql
```

Migrations are applied in alphabetical order. They are **not** tracked automatically — for a real project, use [golang-migrate](https://github.com/golang-migrate/migrate).

---

## Observability

Grafana ships with a pre-built **App Overview** dashboard:
- Requests / sec per service
- 5xx error rate
- p95 latency
- Total request volume

Open **http://localhost:3000** → login `admin` / `admin` (or the password in your `.env`).

---

## CI

Every PR runs `go build ./...` and `go test ./...` via GitHub Actions (`.github/workflows/ci.yml`). Broken builds are blocked from merging.
