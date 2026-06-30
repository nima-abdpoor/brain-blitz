# Refactoring Candidates

## P0 — Critical Bugs (Fix Immediately)

### R-01: Auth adapter returns access token instead of refresh token

**Location**: `adapter/auth/client.go:52`

**Problem**: `GetRefreshToken` calls `client.GetAccessToken` on the gRPC stub. The returned "refresh token" is an access token with a 24h expiry instead of 120h. Clients cannot distinguish access from refresh tokens.

**Impact**: High — broken security model; refresh token rotation does not work.

**Suggested solution**: Change `client.GetAccessToken(ctx, req)` to `client.GetRefreshToken(ctx, req)` and use the `CreateRefreshTokenRequest` proto type.

**Priority**: P0 | **Effort**: Trivial (2-line fix) | **Risk**: None

---

### R-02: Remove debug println of JWT secret key

**Location**: `services/auth_app/service/service.go:62`

**Problem**: `fmt.Println(svc.config.SecretKey, signedString, err)` logs the signing secret on every token creation.

**Impact**: Critical security issue in production.

**Suggested solution**: Delete the line. Add structured logging of the error only if needed.

**Priority**: P0 | **Effort**: Trivial | **Risk**: None

---

## P1 — High Priority

### R-03: Add mutex to Game Service WebSocket connection map

**Location**: `services/game_app/service/service.go` — `IdToConnection` map

**Problem**: `svc.connections` (`map[uint64]net.Conn`) is read and written from multiple goroutines: the WebSocket read loop, `ConsumeMatchCreated`, and `ProcessGameCompletion`. Go maps are not concurrency-safe; concurrent access causes a data race.

**Impact**: High — runtime panic under load.

**Suggested solution**: Replace the plain map with a `sync.RWMutex`-protected wrapper or use `sync.Map`. Define a type:
```go
type ConnectionStore struct {
    mu    sync.RWMutex
    conns map[uint64]net.Conn
}
```

**Priority**: P1 | **Effort**: Small (1 day) | **Risk**: Low

---

### R-04: Fix `UpsertReadyPlayer` premature readiness condition

**Location**: `services/game_app/repository/game.go:236`

**Problem**: `ExpectedPlayers - len(Players) == 1` returns `true` when there is still one player missing. This starts the game prematurely.

**Impact**: High — game starts with only one of two players marked ready.

**Suggested solution**: Change the condition to:
```go
if len(gs.Players) >= gs.ExpectedNumberOfPlayers {
    return true, nil
}
```

**Priority**: P1 | **Effort**: Trivial | **Risk**: Low

---

### R-05: Switch Kafka consumer to consumer groups

**Location**: `adapter/broker/kafka-broker.go:79`

**Problem**: `consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)` — hard-coded to partition 0, no offset commit, no consumer group. Prevents horizontal scaling and loses messages on restart.

**Impact**: High — architectural blocker for production scaling.

**Suggested solution**: Migrate to `sarama.ConsumerGroup` with a configurable group ID. This enables multiple service instances to share load and preserves offset state.

**Priority**: P1 | **Effort**: Medium (2-3 days) | **Risk**: Medium — requires topic/partition management

---

## P2 — Medium Priority

### R-06: Centralize Kafka topic names as constants

**Location**: `services/match_app/service/service.go:143`, `services/game_app/service/consumer.go:44`, `services/question_app/service/consumer.go:49`

**Problem**: Topic strings `"matchMaking_v1_matchUsers"` and `"question_v1_questions"` are inline strings. Only `GAME_V1_JOIN_MATCH_QUEUE_REQUESTED` is a constant in `contract/event/events.go`.

**Impact**: Medium — typos cause silent failures; no single source of truth.

**Suggested solution**: Add all topic names to `contract/event/events.go`:
```go
const (
    GAME_V1_JOIN_MATCH_QUEUE_REQUESTED = "GAME_V1_JOIN_MATCH_QUEUE_REQUESTED"
    MATCH_V1_MATCH_USERS               = "matchMaking_v1_matchUsers"
    QUESTION_V1_QUESTIONS              = "question_v1_questions"
)
```

**Priority**: P2 | **Effort**: Trivial | **Risk**: None

---

### R-07: Decompose Game Service `service.go`

**Location**: `services/game_app/service/service.go` (689 lines)

**Problem**: A single file handles: WebSocket command routing, match event consumption, question event consumption, game state management, player answer validation, leaderboard queries, and task scheduling coordination.

**Impact**: Medium — hard to test individual concerns; high cognitive load.

**Suggested solution**: Extract into focused files:
- `websocket_handler.go` — `readMessage`, `writeMessage`, WebSocket command dispatch
- `match_consumer.go` — `ConsumeMatchCreated`
- `question_consumer.go` — `ConsumeQuestions`
- `game_lifecycle.go` — `ProcessGame`, `saveGameStatus`, `getUsersGameStatus`
- `answer_processor.go` — `savePlayerAnswer`, `sendQuestionToPlayer`
- `game_completion.go` — `ProcessGameCompletion`

**Priority**: P2 | **Effort**: Medium (2 days) | **Risk**: Low (purely organizational)

---

### R-08: Implement `AddQuestion` API

**Location**: `services/question_app/service/service.go:32`, `services/question_app/delivery/http/handler.go:24`

**Problem**: `AddQuestion` returns an empty response with no implementation. There is no way to add questions to the database via the API.

**Impact**: Medium — questions must be seeded directly in the database.

**Suggested solution**: Implement `repository.InsertQuestion(ctx, question)` in the question repository and connect it to the service and handler.

**Priority**: P2 | **Effort**: Small (1 day) | **Risk**: Low

---

### R-09: Use `GetProperQuestions` for history-aware question selection

**Location**: `services/question_app/service/service.go:53`

**Problem**: `GetRandomQuestions` is used instead of `GetProperQuestions` which filters out questions already seen by participating users within 30 days. The `user_question_history` table is never populated.

**Impact**: Medium — players may see the same questions repeatedly.

**Suggested solution**:
1. Call `GetProperQuestions` instead of `GetRandomQuestions` in `ConsumeMatchCreated`
2. After selecting questions, insert records into `user_question_history`

**Priority**: P2 | **Effort**: Small (1 day) | **Risk**: Low

---

### R-10: Add startup retry for PostgreSQL connections

**Location**: `cmd/user/main.go:42`, `cmd/question/main.go:41`

**Problem**: If PostgreSQL is unavailable at startup, the service exits with a fatal error. In Docker and Kubernetes, database startup ordering is not guaranteed.

**Suggested solution**: Implement exponential backoff retry (e.g. 5 attempts, 2s initial delay) in `pkg/postgresql.Connect`.

**Priority**: P2 | **Effort**: Small | **Risk**: Low

---

### R-11: Add Prometheus metrics

**Location**: Multiple — `// todo add metrics` in 15+ locations

**Problem**: No application-level metrics (request latency, error rates, queue depths, active games, matchmaking wait time).

**Suggested solution**: Integrate `prometheus/client_golang`. Add metrics at:
- HTTP middleware (request count, latency, error rate per endpoint)
- Kafka consumer (messages processed, processing time, errors)
- Match scheduler (queue depth per category, match rate)
- Game service (active WebSocket connections, answers per second, game completion rate)

**Priority**: P2 | **Effort**: Large (3-5 days) | **Risk**: Low

---

## P3 — Lower Priority

### R-12: Add Kubernetes deployment for Question Service

**Location**: `infra/kubernetes/deployment/`

**Problem**: No K8s manifest for Question Service.

**Suggested solution**: Create `infra/kubernetes/deployment/question.yaml` and `infra/kubernetes/volume/question-cm.yaml` following the same pattern as other services.

**Priority**: P3 | **Effort**: Trivial | **Risk**: None

---

### R-13: Consolidate duplicate `Category` and `Difficulty` type definitions

**Location**: `services/match_app/service/entity.go`, `services/game_app/service/entity.go`, `services/question_app/service/entity.go`

**Problem**: `Category` and `Difficulty` types and their map functions are duplicated across three services with slightly different implementations (e.g. Question Service's `MapToCategory` defaults to `MUSIC` for unknown values; Game Service defaults to `UNKNOWN`).

**Suggested solution**: Extract to a shared `pkg/domain` or `contract/domain` package. Each service can import from there. Note: this requires careful consideration of service independence.

**Priority**: P3 | **Effort**: Medium | **Risk**: Medium — crosses service boundaries

---

### R-14: Remove dead `GRPCServer` config from Match Service

**Location**: `services/match_app/config.go:9`

**Problem**: `GRPCServer grpc.Config` is in the config struct but no gRPC server is wired in `app.go`.

**Suggested solution**: Remove the field (or document it as planned future work).

**Priority**: P3 | **Effort**: Trivial | **Risk**: None

---

### R-15: Update Game document status on game completion

**Location**: `services/game_app/service/service.go:ProcessGameCompletion`

**Problem**: After the game ends, the MongoDB `game` document's `status` field is never updated to `FINISHED`.

**Suggested solution**: After sending the final leaderboard, call a `repository.UpdateGameStatus(ctx, gameId, GameStatusFinished)` method.

**Priority**: P3 | **Effort**: Small | **Risk**: Low
