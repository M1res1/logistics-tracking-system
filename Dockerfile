# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

# SERVICE build arg — matches the folder name under ./services/
ARG SERVICE
ENV SERVICE=${SERVICE}

WORKDIR /app

# Cache dependencies before copying source
COPY go.mod go.sum ./
RUN go mod download

# Copy full source, then build only the target service
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/service ./services/${SERVICE}/cmd/...

# ── Final stage ───────────────────────────────────────────────────────────────
FROM alpine:3.19

# Security: run as non-root
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /bin/service /app/service

USER app

# PORT is injected via docker-compose environment
EXPOSE ${PORT}

ENTRYPOINT ["/app/service"]
