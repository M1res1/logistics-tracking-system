# FoodApp — Logistics Tracking System

Full-stack food delivery platform with microservices backend and React frontend.

## Quick Start

```bash
# 1. Clone
git clone https://github.com/M1res1/logistics-tracking-system.git
cd logistics-tracking-system

# 2. Boot everything (first run builds all images ~3-5 min)
docker compose up --build -d

# 3. Open the app
open http://localhost:3000
```

That's it. No manual config needed — `.env` defaults work out of the box.

---

## What's Running

| Service        | URL                        | Description                    |
|----------------|----------------------------|--------------------------------|
| **Frontend**   | http://localhost:3000      | React app (Customer/Driver/Restaurant views) |
| **API Gateway**| http://localhost:80        | Nginx — routes `/api/v1/*` to services |
| Auth           | :8080 (internal)           | Register, login, JWT           |
| Orders         | :8081 (internal)           | Order state machine            |
| Deliveries     | :8082 (internal)           | Assignment, GPS tracking       |
| Restaurants    | :8083 (internal)           | Menu management, geo-search    |
| Payments       | :8084 (internal)           | Wallet, idempotent payments    |
| Prometheus     | http://localhost:9090      | Metrics                        |
| Grafana        | http://localhost:3001      | Dashboards (admin / admin)     |

---

## User Roles

Register with one of three roles:

| Role | What you can do |
|------|----------------|
| `CUSTOMER` | Browse restaurants, place orders, track delivery |
| `DRIVER` | Accept deliveries, update GPS location |
| `RESTAURANT` | Manage menu, accept/reject kitchen orders |

---

## Project Structure

```
.
├── docker-compose.yml          # Boots everything with one command
├── .env                        # Default env vars (safe for local dev)
├── nginx/nginx.conf            # API gateway routing
├── frontend/                   # React + TypeScript app
│   ├── src/
│   │   ├── pages/              # CustomerPage, DriverPage, RestaurantPage
│   │   ├── api/                # Typed API clients
│   │   └── contexts/           # AuthContext (JWT, user)
│   └── Dockerfile
├── backend/
│   ├── Dockerfile              # Multi-stage Go build (order/delivery/restaurant/payment)
│   ├── pkg/                    # Shared: config, database, middleware, response, kafka
│   └── services/
│       ├── auth-go/            # Auth service (own go.mod) — :8080
│       ├── order/              # Order service — :8081
│       ├── delivery/           # Delivery + GPS tracking — :8082
│       ├── restaurant/         # Restaurant + menu — :8083
│       └── payment/            # Payments + wallet — :8084
└── Makefile                    # make up / down / test / migrate
```

---

## Makefile Commands

```bash
make up          # Start postgres, redis, kafka
make down        # Stop all containers
make migrate     # Run SQL migrations from backend/migrations/
make test        # Run Go tests (main module)
make test-auth   # Run auth-go tests
make test-all    # Run all tests
make logs        # Tail all logs  (make logs s=auth to filter)
```

---

## Development (without Docker)

```bash
# Start infra only
make up

# Run a service locally
make run-auth
make run-order
make run-deliveries
make run-restaurants
make run-payments

# Frontend dev server
cd frontend && npm install && npm run dev
# → http://localhost:5173  (change VITE_API_BASE_URL if needed)
```

---

## CI

Every PR runs `go build ./...` and `go test ./...` for both Go modules via GitHub Actions.
