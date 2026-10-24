# ── Stage 1: Build frontend ─────────────────────────────────────
FROM node:22-alpine AS frontend-build
WORKDIR /app/web

# VITE_* values are baked into the JS bundle at build time. Pass them in via
# `--build-arg` (or Coolify's "Build Args" panel — *not* the runtime env panel).
ARG VITE_SENTRY_DSN
ENV VITE_SENTRY_DSN=$VITE_SENTRY_DSN

COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
RUN npm run build

# ── Stage 2: Build backend ──────────────────────────────────────
FROM golang:1.25-alpine AS backend-build
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Embed the frontend dist into the binary's working dir
COPY --from=frontend-build /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /diagram ./cmd/server

# ── Stage 3: Production image ──────────────────────────────────
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata wget
RUN adduser -D -u 1000 diagram

WORKDIR /app
COPY --from=backend-build /diagram .
COPY --from=frontend-build /app/web/dist ./web/dist

USER diagram

EXPOSE 8081

ENV APP_ENV=production
ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=8081

# Runtime healthcheck for platforms that drive `docker run` directly
# (compose has its own block; this is a belt-and-braces fallback).
HEALTHCHECK --interval=15s --timeout=5s --start-period=15s --retries=5 \
    CMD wget -qO- http://localhost:8081/health || exit 1

ENTRYPOINT ["./diagram"]
