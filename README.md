# logistics-tracking-system
logistics-tracking-system web app
# 🍔 Food Delivery Logistics System

> A microservices backend for a food delivery platform — think Uber Eats, built in Go.

Built as a university project. Connects customers, restaurants, and delivery drivers through 5 independent Go services.

---

## What It Does
The Food Delivery Logistics System is a backend platform that connects customers, restaurants, and delivery drivers — similar to how Uber Eats or Bolt Food works under the hood. Customers can browse restaurant menus, place orders, and track their delivery driver's location in real time. Restaurants receive orders digitally, manage their menu, and control their kitchen flow by accepting or rejecting incoming orders. A delivery driver gets automatically assigned the moment the food is ready, and their GPS position is updated live so the customer always knows where their order is. The entire system is built in Go using a microservices architecture, with each core function — authentication, orders, delivery, restaurants, and payments — running as an independent service.

A customer browses restaurants, places an order, pays, and tracks their driver in real time. The restaurant receives the order digitally, confirms it, and marks it ready. A driver gets auto-assigned, picks it up, and delivers it. All coordinated through this backend.

---

## Services

| Service | Port | What it does |
|---|---|---|
| **Auth** | `:8080` | Register, login, JWT tokens, role middleware |
| **Order** | `:8081` | Create orders, state machine, status tracking |
| **Delivery** | `:8082` | Assign drivers, GPS tracking, ETA calculation |
| **Restaurant** | `:8083` | Menu management, accept/reject orders |
| **Payment** | `:8084` | Process payments, wallet, refunds |

All routes go through **Nginx** on `http://localhost/api/v1/...`

---

## Tech Stack

- **Go 1.21+** — all services
- **Gin** — HTTP router
- **GORM** — ORM / database access
- **PostgreSQL** — primary database
- **Redis** — token blacklist, caching, idempotency keys
- **Kafka** — async events between services
- **Docker + Compose** — local infrastructure
- **Nginx** — reverse proxy / API gateway
- **GitHub Actions** — CI on every PR

---

## Getting Started

### Prerequisites

- [Docker](https://www.docker.com/) & Docker Compose
- [Go 1.21+](https://go.dev/dl/)
- Git

### Run locally

```bash
# 1. Clone the repo
git clone https://github.com/your-team/food-delivery.git
cd food-delivery

# 2. Set up environment
cp .env.example .env

# 3. Start infrastructure (Postgres, Redis, Kafka)
make up

# 4. Run database migrations
make migrate

# 5. Start all services
make run-all
```

All APIs are now available at `http://localhost/api/v1/`

---

## Makefile Commands

```bash
make up          # start postgres, redis, kafka
make down        # stop all containers
make migrate     # run SQL migrations
make run-all     # start all 5 Go services
make test        # run all tests
make build-all   # compile all services to /bin
```

---

## Project Structure

```
food-delivery/
│
├── pkg/                        # shared code imported by all services
│   ├── middleware/auth.go      # JWT validation middleware
│   ├── response/response.go    # standard JSON response helpers
│   ├── config/config.go        # load .env into Config struct
│   ├── database/postgres.go    # open DB connection
│   └── kafka/                  # producer & consumer helpers
│
├── services/
│   ├── auth/                   # :8080
│   ├── order/                  # :8081
│   ├── delivery/               # :8082
│   ├── restaurant/             # :8083
│   └── payment/                # :8084
│
├── migrations/                 # versioned SQL files
│   ├── 001_users.sql
│   ├── 002_restaurants.sql
│   ├── 003_menu_items.sql
│   ├── 004_orders.sql
│   ├── 005_order_items.sql
│   ├── 006_deliveries.sql
│   └── 007_payments.sql
│
├── docker-compose.yml
├── .env.example
├── Makefile
└── README.md
```

Each service follows the same internal layout:

```
services/order/
├── main.go
├── handler/
├── service/
├── repository/
└── model/
```

---

## API Overview

### Auth
```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/logout
POST   /api/v1/auth/refresh-token
GET    /api/v1/auth/me
```

### Orders
```
POST   /api/v1/orders
GET    /api/v1/orders/:id
GET    /api/v1/orders/my
POST   /api/v1/orders/:id/cancel
```

### Delivery
```
POST   /api/v1/deliveries/assign
POST   /api/v1/deliveries/:id/accept
POST   /api/v1/deliveries/:id/pickup
POST   /api/v1/deliveries/:id/complete
PUT    /api/v1/deliveries/:id/location
GET    /api/v1/deliveries/:id/location
```

### Restaurants
```
GET    /api/v1/restaurants
POST   /api/v1/restaurants
GET    /api/v1/restaurants/:id/menu
POST   /api/v1/restaurants/:id/menu-items
PUT    /api/v1/restaurants/:id/toggle
POST   /api/v1/restaurants/:id/orders/:orderId/accept
POST   /api/v1/restaurants/:id/orders/:orderId/ready
POST   /api/v1/restaurants/:id/orders/:orderId/reject
```

### Payments
```
POST   /api/v1/payments/process
GET    /api/v1/payments/:id
POST   /api/v1/payments/:id/refund
GET    /api/v1/wallet/:userId
POST   /api/v1/wallet/:userId/topup
```

---

## Order Status Flow

```
PENDING → CONFIRMED → PREPARING → READY → ASSIGNED → IN_TRANSIT → DELIVERED
                                                  ↓
                                             CANCELLED
```

---

## Environment Variables

Copy `.env.example` to `.env` and fill in:

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=fooddelivery
DB_USER=admin
DB_PASS=secret

REDIS_URL=localhost:6379

KAFKA_BROKERS=localhost:9092

JWT_SECRET=your-secret-key-here

AUTH_PORT=8080
ORDER_PORT=8081
DELIVERY_PORT=8082
RESTAURANT_PORT=8083
PAYMENT_PORT=8084
```

---

## Team

| Developer | Service | 
|---|---|
| Dev 1 | Auth Service |
| Dev 2 | Order Service |
| Dev 3 | Delivery Service |
| Dev 4 | Restaurant Service |
| Dev 5 | Payment Service + shared `pkg/` |
| DevOps | Docker, Nginx, CI/CD, Makefile |

> **Note for the team:** Dev 5 sets up `pkg/` and Docker Compose on Day 1. Dev 1 finishes the auth middleware before others protect their routes. These two tasks block everyone else.

---

## Running Tests

```bash
# all services
make test

# single service
cd services/order && go test ./...
```

---

*University Project · Go · May 2026*