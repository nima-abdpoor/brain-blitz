# 01 — Project Overview

## Purpose

BrainBlitz is a real-time multiplayer quiz game backend. Its core value proposition is matching anonymous online players by topic category and running a timed, scored quiz session entirely over WebSocket — with matchmaking, question delivery, answer scoring, and a leaderboard all handled server-side.

## Goals

| Goal | How it is achieved |
|---|---|
| Real-time gameplay | Raw WebSocket via `gobwas/ws` in Game Service |
| Fair matchmaking | Redis sorted sets (by join timestamp) with periodic scheduler |
| Stateless authentication | JWT tokens signed with HMAC-SHA256, validated via ForwardAuth middleware in Traefik |
| Horizontal scalability (infrastructure level) | Independent microservices, Kafka for async decoupling, MongoDB replica set |
| Question freshness | User question history tracked; seen questions excluded within 30 days |
| Pluggable configuration | YAML + env-var overrides via `knadh/koanf` |

## Major Components

| Component | Role |
|---|---|
| **Auth Service** | Mints and validates HS256 JWT tokens; only stateless token operations, no database |
| **User Service** | User registration, authentication (via Auth gRPC), and profile retrieval; stores in PostgreSQL |
| **Match Service** | Holds players in Redis waiting lists per category; scheduler pairs players and publishes match events to Kafka |
| **Game Service** | WebSocket hub; creates game sessions in MongoDB; consumes match and question events; enforces answer timing; calculates scores |
| **Question Service** | Stores questions in PostgreSQL; selects questions per match and publishes them to Kafka |
| **Traefik** | API gateway; route matching, prefix stripping, ForwardAuth against Auth Service |
| **Kafka** | Async event bus; decouples Game ↔ Match ↔ Question |
| **Redis** | Matchmaking queues (sorted sets), game state cache, Asynq task queue |
| **MongoDB** | Persistent game documents, player answer records, leaderboard aggregation |
| **PostgreSQL** | User credentials and question bank |

## Architecture Style

**Microservices** with a mix of communication patterns:

- **Synchronous HTTP** — public-facing CRUD (signup, login, profile, add-to-waiting-list)
- **Synchronous gRPC** — inter-service token operations (User calls Auth; internal gRPC server on User for future use)
- **Asynchronous Kafka** — game flow events (join queue → match created → questions published)
- **WebSocket** — full-duplex real-time game session between client and Game Service

## Design Philosophy

1. **Layer separation** — each service has `delivery/` (HTTP/gRPC/WS handlers), `service/` (business logic), and `repository/` (data access). These layers are decoupled via interfaces.
2. **Composition at `app.go`** — wiring of dependencies happens in each service's `app.go`, not inside business logic.
3. **Uniform error handling** — `pkg/err_app` defines `AppError` with both HTTP status and gRPC code; `ToHTTPJson` / `ToGRPCJson` convert at the delivery boundary.
4. **Config-first** — all timeouts, expiry durations, ports, and feature parameters are in config structs loaded from YAML, overridable by env vars.
5. **Graceful shutdown** — every service handles `SIGTERM`/`SIGINT` with a configurable total shutdown timeout and per-server shutdown contexts.
