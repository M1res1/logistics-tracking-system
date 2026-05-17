# Build stage
FROM golang:1.25-alpine AS builder

ARG SERVICE
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/service ./services/${SERVICE}

# Final stage
FROM alpine:3.19
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /bin/service /app/service
USER app
ENTRYPOINT ["/app/service"]
