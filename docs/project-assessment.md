# Project Assessment

## Executive Summary

BrainBlitz is a well-structured real-time multiplayer quiz game backend implemented as Go microservices. The architecture correctly separates concerns across five services and demonstrates good understanding of asynchronous event-driven communication. The developer experience is solid: the project runs with a single `docker-compose up` command and CI validates every push to `develop`.

However, the project has **two critical bugs** (auth adapter returns wrong token type; JWT secret key printed to stdout), **one data race** (concurrent map access in Game Service), and numerous unimplemented features (metrics, AddQuestion, game status updates, token refresh). These issues must be resolved before the system can be considered production-ready.

---

## Architecture Review

**Strengths**:
- Clean three-layer structure (`delivery → service → repository`) consistently applied across all services
- Dependency inversion via `Repository` and `Broker` interfaces enables testability
- Traefik ForwardAuth correctly externalizes authentication from downstream services
- Kafka decouples Match, Game, and Question services; failure in one does not block others
- MongoDB replica set and Kafka provide data durability for game events
- Graceful shutdown pattern is implemented consistently in all services

**Weaknesses**:
- Kafka consumer uses low-level partition consumer (partition 0 only, no consumer groups) — horizontal scaling is impossible
- Game Service `connections` map is not concurrency-safe; data race under load
- WebSocket connections are in-memory; a second Game Service replica cannot send messages to players connected to the first replica
- Auth adapter `GetRefreshToken` calls the wrong gRPC method (calls `GetAccessToken`)
- No application-level metrics or distributed tracing

---

## Business Review

The game loop (join → wait → match → question delivery → answer → leaderboard → game over) is substantially implemented and working. The matchmaking logic correctly handles FIFO ordering and category grouping.

Key business features that are missing or incomplete:
- No way to add questions via the API (AddQuestion is a stub)
- Refresh token flow is broken (returns access token instead)
- No token revocation or logout
- History-aware question selection code exists but is never called
- Game completion does not update game document status

---

## Strengths

1. **Clear service decomposition** — Auth, User, Match, Game, Question each have a single responsibility that is easy to explain.
2. **Consistent code structure** — Any Go developer can navigate any service without learning a new pattern.
3. **Protobuf contracts** — Kafka messages are strongly typed and versioned via `.proto` files.
4. **Configuration flexibility** — YAML defaults + env var overrides work correctly via `koanf`.
5. **Unified error type** — `AppError` with both HTTP and gRPC status codes avoids duplication at the delivery layer.
6. **Graceful shutdown** — All services handle OS signals correctly.
7. **MongoDB replica set** — Game data is replicated with automatic failover.
8. **Asynq deferred tasks** — Game completion timer is resilient to server restarts (Redis-backed).
9. **Docker Compose** — One-command local environment including monitoring.
10. **GitHub Actions CI** — Build + test gate on every push.

---

## Weaknesses

1. **Critical auth bug**: Refresh token returns access token (auth adapter).
2. **Security leak**: JWT signing secret printed to stdout.
3. **Data race**: Game Service WebSocket connection map has no synchronization.
4. **Premature game start**: UpsertReadyPlayer readiness check is off by one.
5. **Not horizontally scalable**: Kafka single-partition consumers; in-memory WebSocket map.
6. **Zero application metrics**: No Prometheus instrumentation despite 15+ todo comments.
7. **AddQuestion is a stub**: No operational way to manage the question bank.
8. **Question freshness broken**: `GetProperQuestions` exists but is not called; history table is never populated.
9. **No reconnection handling**: If a WebSocket client disconnects mid-game, there is no resume flow.
10. **Kubernetes incomplete**: Question Service has no K8s manifests.

---

## Technical Debt

| Item | Severity | File |
|---|---|---|
| Auth adapter GetRefreshToken bug | Critical | `adapter/auth/client.go` |
| fmt.Println of secret key | Critical | `services/auth_app/service/service.go:62` |
| No concurrency safety on connections map | High | `services/game_app/service/service.go` |
| UpsertReadyPlayer ready condition | High | `services/game_app/repository/game.go:236` |
| No Kafka consumer groups | High | `adapter/broker/kafka-broker.go` |
| AddQuestion not implemented | Medium | `services/question_app/service/service.go` |
| No metrics anywhere | Medium | 15+ locations |
| GetProperQuestions never called | Medium | `services/question_app/service/service.go` |
| Game status not updated on completion | Medium | `services/game_app/service/service.go` |
| No PostgreSQL startup retry | Medium | `cmd/user/main.go`, `cmd/question/main.go` |
| Dead GRPCServer config in Match | Low | `services/match_app/config.go` |
| Inline Kafka topic strings | Low | Multiple service files |
| Duplicate Category/Difficulty types | Low | Three service packages |
| Dead sqlc Makefile target | Low | `Makefile` |
| Question Service missing K8s manifest | Low | `infra/kubernetes/deployment/` |

---

## Security Concerns

| Concern | Severity | Notes |
|---|---|---|
| JWT secret key logged to stdout | Critical | `services/auth_app/service/service.go:62` |
| JWT secret default is literal `SECRET_KEY` | High | No protection if env var not set |
| Redis runs without auth in docker-compose | Medium | `--protected-mode no`, no password |
| Traefik dashboard is public (`api.insecure=true`) | Medium | Exposes internal routing info |
| CORS is `allow_origins: "*"` | Medium | All services accept requests from any origin |
| MongoDB exporter exposed publicly | Medium | All metrics accessible without auth |
| No HTTPS/TLS configured | Medium | All traffic is plaintext |
| No rate limiting | Medium | No protection against brute force or abuse |

---

## Performance Concerns

| Concern | Impact | Notes |
|---|---|---|
| Kafka single-partition consumer | High | Cannot distribute load across instances |
| In-memory WebSocket map | High | Single-instance bottleneck for Game Service |
| MongoDB leaderboard query loads all player answers | Medium | `GetLeaderBoard` fetches all answers then aggregates in Go; could use MongoDB aggregation pipeline |
| No Redis pipelining | Low | Multiple Redis commands issued sequentially |
| `ORDER BY RANDOM()` in PostgreSQL question query | Low | Full table scan; acceptable for small question banks |

---

## Testing Assessment

| Area | Coverage | Notes |
|---|---|---|
| Auth service token logic | Good | Unit tests exist in `service/service_test.go` |
| Match service validation | Good | Unit tests for validators |
| User service logic | Partial | Some unit tests; no full service test |
| User repository | Partial | sqlmock-based test |
| Game service | None | Largest and most complex service has no tests |
| Question service | None | No tests |
| Adapters (Kafka, Redis, WS, Asynq) | None | No tests |
| Integration tests | Minimal | 3 HTTP integration tests (auth, match, user login) |
| End-to-end game flow | None | No automated test of full game session |

---

## Documentation Assessment

**Before this assessment**:
- `README.md` was minimal (emoji-heavy, missing setup steps, missing configuration reference)
- No architecture documentation beyond a single PNG image
- No developer guides

**After this assessment**:
- Comprehensive README with all sections
- `docs/context/` — 8 AI-optimized context documents
- `docs/refactoring/` — prioritized refactoring roadmap
- `docs/business/` — domain model, flows, feature map, roadmap, use cases
- `docs/diagrams.md` — Mermaid diagrams for architecture, Kafka flow, state machine, deployment
- `docs/project-assessment.md` — this document

---

## Recommended Priorities

### Immediate (this sprint)
1. Fix auth adapter `GetRefreshToken` to call the correct gRPC method
2. Remove `fmt.Println(svc.config.SecretKey, ...)` from auth service
3. Add `sync.RWMutex` to Game Service WebSocket connection map
4. Fix `UpsertReadyPlayer` ready condition (off by one)

### Short Term (next 2 sprints)
5. Implement `AddQuestion` API end-to-end
6. Activate `GetProperQuestions` and populate question history
7. Update game document status to `FINISHED` on completion
8. Add Prometheus metrics to HTTP middleware across all services
9. Add Kubernetes manifest for Question Service
10. Centralize Kafka topic strings in `contract/event/events.go`

### Medium Term (next quarter)
11. Migrate Kafka consumers to consumer groups
12. Add startup retry for PostgreSQL
13. Implement distributed WebSocket fan-out (Redis Pub/Sub) for Game Service scaling
14. Implement token refresh endpoint
15. Add structured end-to-end integration tests for the full game flow

---

## Long-Term Recommendations

1. **Extract a game engine package** — the Game Service's `service.go` (689 lines) should be decomposed into focused modules: WebSocket protocol, game lifecycle, answer processing, event consumption.

2. **Introduce distributed tracing** — once metrics are in place, add OpenTelemetry traces to correlate a user action (WebSocket command) through Kafka events across three services.

3. **Consider a dedicated WebSocket gateway** — as player count grows, separating WebSocket connection management from game logic allows independent scaling of the connection layer.

4. **Harden production configuration** — set Redis passwords, disable Traefik insecure mode, enforce CORS, add TLS termination, and use Kubernetes Secrets for all sensitive values.

5. **Question bank administration** — implement a full admin API for question CRUD with category and difficulty filtering, protected by the `admin` role.
