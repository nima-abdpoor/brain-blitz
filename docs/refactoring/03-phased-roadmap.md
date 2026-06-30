# Phased Refactoring Roadmap

## Phase 1 — Critical Bug Fixes

**Goals**: Eliminate security issues and prevent runtime panics.

**Expected outcomes**: Production-safe auth flow; Game Service stable under concurrent load.

**Files involved**:
- `adapter/auth/client.go`
- `services/auth_app/service/service.go`
- `services/game_app/service/service.go` (connection map)
- `services/game_app/repository/game.go` (ready condition)

**Dependencies**: None — fixes are isolated.

**Risk level**: Low

**Estimated complexity**: 1–2 days

**Tasks**:
- R-01: Fix `GetRefreshToken` to call the correct gRPC method
- R-02: Remove `fmt.Println` of secret key
- R-03: Add `sync.RWMutex` wrapper around `connections` map in Game Service
- R-04: Fix `UpsertReadyPlayer` readiness condition from `== 1` to `>= expected`

---

## Phase 2 — Code Cleanup

**Goals**: Eliminate dead code, fix configuration inconsistencies, centralize shared constants.

**Expected outcomes**: Cleaner codebase with no confusion over dead features; all Kafka topics defined in one place.

**Files involved**:
- `contract/event/events.go`
- `services/match_app/service/service.go`
- `services/game_app/service/consumer.go`
- `services/question_app/service/consumer.go`
- `services/match_app/config.go`
- `services/auth_app/service/service.go` (remove `// todo` blocks)
- `Makefile` (remove dead sqlc target)

**Dependencies**: Phase 1 complete

**Risk level**: Low

**Estimated complexity**: 1–2 days

**Tasks**:
- R-06: Centralize all Kafka topic names as constants in `contract/event/events.go`
- R-14: Remove dead `GRPCServer` config from Match Service
- Clean up `// todo add metrics` comments (replace with structured log or leave a GitHub issue)
- Remove the dead `sqlc-generate` Makefile target
- Document the `question_app` `users` and `match_questions` tables as unused in migrations

---

## Phase 3 — Architecture Improvements

**Goals**: Fix the Kafka consumer architecture to enable horizontal scaling. Implement missing core features (AddQuestion, game status update, history-aware questions).

**Expected outcomes**: Services are horizontally scalable; question freshness works; questions can be administered via API.

**Files involved**:
- `adapter/broker/kafka-broker.go`
- `services/question_app/service/service.go`
- `services/question_app/repository/questions.go`
- `services/game_app/service/service.go`
- `services/game_app/repository/game.go`
- `services/question_app/delivery/http/handler.go`
- `infra/kubernetes/deployment/question.yaml` (new file)

**Dependencies**: Phase 2 complete; Kafka topic constants from Phase 2 used here.

**Risk level**: Medium — Kafka consumer group migration may require topic recreation.

**Estimated complexity**: 5–7 days

**Tasks**:
- R-05: Migrate to Kafka consumer groups (`sarama.ConsumerGroup`)
- R-08: Implement `AddQuestion` in Question Service (repository + service + handler)
- R-09: Use `GetProperQuestions` and populate `user_question_history`
- R-15: Update game document status to `FINISHED` on game completion
- R-12: Add Kubernetes deployment manifest for Question Service

---

## Phase 4 — Performance Improvements

**Goals**: Improve WebSocket scalability; add startup resilience.

**Expected outcomes**: Game Service can run as multiple replicas; services survive transient database unavailability at startup.

**Files involved**:
- `services/game_app/service/service.go` (WebSocket hub redesign)
- `pkg/postgresql/db.go`
- `services/game_app/app.go`
- Possibly introduce Redis Pub/Sub for WebSocket fan-out

**Dependencies**: Phase 3 complete (consumer groups required first).

**Risk level**: High — WebSocket hub redesign is a significant architectural change.

**Estimated complexity**: 5–10 days

**Tasks**:
- R-10: Add retry with backoff for PostgreSQL connections
- Design and implement a distributed WebSocket fan-out using Redis Pub/Sub (when player receives a match event, the Game Service instance holding the connection is notified via Redis)
- Replace in-memory `connections` map with event-driven fan-out to enable multiple Game Service replicas
- Load test the new design before merging

---

## Phase 5 — Developer Experience

**Goals**: Improve observability and code organization.

**Expected outcomes**: Prometheus metrics available for all services; Game Service business logic is easier to navigate and test.

**Files involved**:
- `services/game_app/service/service.go` → multiple focused files
- All service HTTP middleware files
- New files: `pkg/metrics/` or per-service metrics registration
- `prometheus.yml`

**Dependencies**: Phase 2 complete.

**Risk level**: Low (metrics) to Medium (decomposition).

**Estimated complexity**: 5–7 days

**Tasks**:
- R-07: Decompose `services/game_app/service/service.go` into focused files
- R-11: Add Prometheus metrics to HTTP middleware, Kafka consumers, Game Service, Match scheduler
- Update `prometheus.yml` to scrape all services
- Add Grafana dashboards for key metrics

---

## Phase 6 — Long-Term Modernization

**Goals**: Address fundamental architectural constraints around shared types and production security.

**Expected outcomes**: Shared domain types reduce duplication; production deployment is secure.

**Files involved**:
- New: `pkg/domain/` or `contract/domain/`
- `services/match_app/service/entity.go`
- `services/game_app/service/entity.go`
- `services/question_app/service/entity.go`
- `infra/deploy/*/development/config.yaml` (security config review)
- `docker-compose.yml` (Redis password, Traefik, MongoDB auth)

**Dependencies**: All previous phases.

**Risk level**: Medium–High — shared types cross service boundaries; security hardening requires coordination.

**Estimated complexity**: 5–10 days

**Tasks**:
- R-13: Extract shared `Category` and `Difficulty` types to a common package
- Harden default configuration: remove `allow_origins: "*"`, set Redis password, disable Traefik insecure dashboard in production
- Introduce structured secrets management (K8s Secrets for all sensitive values)
- Add integration test suite covering the full game flow end-to-end
- Review and enforce CORS policy per service
