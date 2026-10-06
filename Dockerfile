# ─────────────────────────────────────────────────────────────────────────────
# OENGO Go Enterprise Backend Daemon Dockerfile
# ─────────────────────────────────────────────────────────────────────────────
# Stage 1: Build binary with Go 1.26
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/oengo-api .

# Stage 2: Minimal runtime image
FROM alpine:3.20

WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata wget

COPY --from=builder /app/oengo-api /app/oengo-api

ENV PORT=3001
EXPOSE 3001

HEALTHCHECK --interval=20s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:3001/health || exit 1

ENTRYPOINT ["/app/oengo-api"]
