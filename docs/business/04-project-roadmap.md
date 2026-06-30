# Project Roadmap

This roadmap is inferred from the current implementation, missing features, and code TODOs. It is not derived from any external product spec.

## Phase 1: Stability & Security (Immediate)

Fix the bugs and security issues that affect the core game experience today.

| Item | Type | Priority |
|---|---|---|
| Fix refresh token returns access token | Bug fix | P0 |
| Remove JWT secret key debug print | Security | P0 |
| Fix `UpsertReadyPlayer` starts game 1 player early | Bug fix | P1 |
| Add mutex to Game Service WebSocket connection map | Bug fix | P1 |
| Add startup retry for PostgreSQL connections | Resilience | P2 |

**Milestone**: BrainBlitz is functionally correct and safe to deploy.

---

## Phase 2: Core Missing Features

Complete the features that are partially implemented or stubbed.

| Item | Type | Notes |
|---|---|---|
| Implement `AddQuestion` API | Feature | Enable question administration without direct DB access |
| Activate history-aware question selection | Feature | Populate `user_question_history`; use `GetProperQuestions` |
| Update game document status to `FINISHED` | Feature | Enables future game history queries |
| Token refresh endpoint | Feature | Client needs to exchange refresh → access token without re-login |
| Question Service Kubernetes deployment | Infra | Add K8s manifest |

**Milestone**: All declared features work end-to-end.

---

## Phase 3: Observability & Operations

Make the system debuggable and monitorable in production.

| Item | Type | Notes |
|---|---|---|
| Prometheus metrics in all services | Feature | Request latency, error rates, Kafka consumer lag, active games, queue depth |
| Grafana dashboards | Feature | Service health, game activity, matchmaking metrics |
| Distributed tracing (OpenTelemetry) | Feature | Trace a request across services |
| Centralize Kafka topic constants | Cleanup | Single source of truth in `contract/event/events.go` |
| Structured error codes in API responses | Feature | Machine-readable error codes for client-side handling |

**Milestone**: On-call engineer can debug production issues without reading code.

---

## Phase 4: Scalability

Enable the platform to handle multiple concurrent games and users.

| Item | Type | Notes |
|---|---|---|
| Kafka consumer groups | Architecture | Required for horizontal scaling of consumers |
| Distributed WebSocket fan-out (Redis Pub/Sub) | Architecture | Enable multiple Game Service replicas |
| Game Service horizontal scaling | Infra | Requires Kafka consumer groups + distributed WebSocket |
| Rate limiting at Traefik | Security/Ops | Protect against abuse |
| Connection timeout management in WebSocket | Reliability | Detect and clean up stale connections |

**Milestone**: System scales to 100+ concurrent games across multiple service instances.

---

## Phase 5: Game Experience Improvements

Features that enhance the player experience.

| Item | Type | Notes |
|---|---|---|
| Medium and Hard difficulty questions | Feature | `GetRandomQuestions` supports it; make category/difficulty selectable per match |
| Player reconnection to ongoing game | Feature | Resume WebSocket session after disconnect |
| Game history API | Feature | Players can view past games and scores |
| Multiple players per match (not just 2) | Feature | Currently UI assumes 2 but matchmaking supports N |
| Category selection per player | Feature | Current implementation already passes category |
| Spectator mode | Feature | Watch a game in progress |
| Question submission by users | Feature | User-generated content (requires admin review flow) |

---

## Phase 6: Platform Expansion

Longer-term features to expand the product.

| Item | Type | Notes |
|---|---|---|
| Tournament mode | Feature | Bracket-style multi-round competitions |
| Leaderboard across games (global rankings) | Feature | Persistent score tracking across matches |
| Social features (friends, challenges) | Feature | Invite-based matchmaking |
| Mobile push notifications | Feature | Match found, game starting |
| Question bank management UI | Feature | Admin interface for question CRUD |
| Multiple languages | Feature | Internationalized questions |
| TLS termination | Security | HTTPS in production |
| Email verification on signup | Security | Verify user email ownership |
| Password reset flow | Feature | Forgot password via email |
