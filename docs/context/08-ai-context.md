# 08 — AI Context

This file is optimized to bring a future AI assistant up to speed on BrainBlitz quickly and accurately.

## Project Summary

BrainBlitz is a **real-time multiplayer quiz game** backend written in **Go 1.23**. It is a monorepo containing **5 microservices**: `auth`, `user`, `match`, `game`, and `question`. Services communicate via **HTTP**, **gRPC**, **Kafka**, and **raw WebSocket**.

Module name: `BrainBlitz.com/game`

## Architecture Summary

```
Traefik (API Gateway + JWT ForwardAuth)
    ├── Auth Service   (port 5000 HTTP, 6000 gRPC) — JWT mint/validate only, no DB
    ├── User Service   (port 5001 HTTP, 6001 gRPC) — PostgreSQL; calls Auth via gRPC
    ├── Match Service  (port 5002 HTTP)            — Redis sorted sets; Kafka producer + consumer; gocron scheduler
    ├── Game Service   (port 5003 HTTP/WS)         — MongoDB + Redis; Kafka consumer + producer; WebSocket hub; Asynq
    └── Question Svc   (port 5004 HTTP)            — PostgreSQL; Kafka consumer + producer
```

## Kafka Topics

| Topic | Producer | Consumers |
|---|---|---|
| `GAME_V1_JOIN_MATCH_QUEUE_REQUESTED` | Game Service (on WebSocket ADD_TO_WAITING_LIST) | Match Service |
| `matchMaking_v1_matchUsers` | Match Service (scheduler) | Game Service, Question Service |
| `question_v1_questions` | Question Service | Game Service |

Topics are constants in `contract/event/events.go` (only `GAME_V1_JOIN_MATCH_QUEUE_REQUESTED` is there; the other two are inline strings — a known inconsistency).

## Package Structure Pattern

Every service follows:
```
services/<name>_app/
  app.go         ← wiring (no logic)
  config.go      ← aggregate config
  delivery/http/ ← Echo handlers + routes
  delivery/grpc/ ← gRPC handlers
  repository/    ← implements service.Repository interface
  service/       ← business logic; defines Repository interface; entity, param, validator files
```

## Important Business Rules

1. Matchmaking requires **2 players minimum**, same category, joined within the last **20 minutes**.
2. The scheduler pairs players every **15 seconds**.
3. **10 questions** per game, difficulty **EASY** only (hardcoded).
4. Players must send `READY` before questions are delivered.
5. Answers must be received **after** question's `ValidAnswerTime` (the deadline), not before. An answer received "too quickly" (before the deadline) is rejected — this implements an anti-cheat window.
6. Scoring: `BaseScore(5) + Bonus(up to 10)`. Bonus scales with remaining time before deadline. Zero for wrong answers or after deadline.
7. Game ends via Asynq deferred task that fires after total TTL = sum of all question deadlines.
8. Token claims embedded in JWT: `id` (user ID as string) and `role` (`"user"` or `"admin"`).
9. Traefik injects `X-User-ID` and `X-User-Role` headers into downstream requests after ForwardAuth.

## Critical Bugs to Know Before Making Changes

1. **Auth adapter `GetRefreshToken` bug** (`adapter/auth/client.go:52`): calls `GetAccessToken` gRPC method instead of `GetRefreshToken`. Refresh tokens are actually access tokens.
2. **Debug print of secret key** (`services/auth_app/service/service.go:62`): `fmt.Println(svc.config.SecretKey, ...)` — remove before production.
3. **Unsafe map in Game Service**: `svc.connections` (`map[uint64]net.Conn`) is accessed from multiple goroutines without synchronization. Risk of data race and panic.
4. **`UpsertReadyPlayer` fires too early** (`services/game_app/repository/game.go:236`): the condition `ExpectedPlayers - readyPlayers == 1` returns `true` when one player is still missing.

## Common Pitfalls

- **Config loading**: env vars use `__` (double underscore) as level separator and a service-specific prefix (`AUTH_`, `USER_`, `GAME_`, `MATCH_`, `QUESTION_`). YAML uses dot notation.
- **Kafka consumer is single-partition, no consumer groups**: `adapter/broker/kafka-broker.go` uses low-level `ConsumePartition(topic, 0, sarama.OffsetNewest)`. Horizontal scaling of consumers will cause duplicate processing.
- **WebSocket connections are in-memory**: scaling Game Service to multiple instances will break — a player connected to instance A cannot receive messages published to instance B's connection map.
- **MongoDB replica set required**: Game Service will not work with a standalone MongoDB instance.
- **Question Service missing K8s manifest**: do not assume question service is deployed in Kubernetes; a manifest needs to be created.
- **`AddQuestion` is a stub**: the Question Service has no working API to add questions. Questions must be seeded directly into PostgreSQL.
- **Auth gRPC server on User Service**: `services/user_app/delivery/grpc/` implements a gRPC server that is started (port 6001) but its purpose is not fully defined in the current codebase.

## Package Responsibilities Summary

| Path | What it does |
|---|---|
| `pkg/err_app` | Defines `AppError` with HTTP+gRPC status; `Wrap`/`New`/`ToHTTPJson`/`ToGRPCJson` |
| `pkg/cfg_loader` | Loads YAML + env vars into a typed config struct via koanf |
| `pkg/logger` | Global slog+lumberjack JSON logger; call `logger.Init(cfg)` then `logger.New()` |
| `pkg/postgresql` | Opens `*sql.DB` connection to PostgreSQL |
| `pkg/mongo` | Opens MongoDB client for a replica set cluster |
| `pkg/postgresqlmigrator` | Runs SQL files from a directory as up/down migrations |
| `pkg/cache_manager` | Thin Redis cache wrapper (set/get/delete/TTL) |
| `adapter/broker` | Kafka producer+consumer with retry; implements `Broker` interface |
| `adapter/redis` | Redis sorted set + string operations; wraps `go-redis` |
| `adapter/auth` | gRPC client that calls Auth Service to get tokens; implements `TokenClient` |
| `adapter/websocket` | Wraps `gobwas/ws` for raw WebSocket upgrade, read, write |
| `adapter/task-queue` | Asynq publisher (enqueue) and worker (process) with Redis backend |
| `contract/event` | Kafka topic name constants |
| `contract/auth/golang` | Generated gRPC stubs for Auth Service |
| `contract/match/golang` | Generated proto types for match events |
| `contract/question/golang` | Generated proto types for question events |

## Database Schema Quick Reference

### PostgreSQL: user_db

```sql
users (id SERIAL PK, username UNIQUE, display_name, role, password, created_at, updated_at)
```

### PostgreSQL: question_db

```sql
questions (id UUID PK, content, correct_answer, choices TEXT[], category, difficulty, created_at)
user_question_history (user_id BIGINT, question_id UUID, seen_at) PK(user_id, question_id)
match_questions (match_id UUID, question_id UUID) PK(match_id, question_id)
users (id UUID PK, username UNIQUE)  -- unused, exists from migration
```

### MongoDB: BB-game

```
game           { _id, players[], match_id, category[], status, questions[], created_at, updated_at }
player_answers { game_id, question_id, player_id, player_choice, correct_choice,
                 answer_time, valid_time_to_answer, time_diff, Option[], point, category }
```

### Redis key patterns

```
waiting_users:<CATEGORY>      — sorted set, Match Service waiting list
game_questions_<gameId>       — JSON GameQuestions, Game Service cache
game_questions_<matchId>      — JSON GameQuestions (before gameId is known)
game_user_status_<userId>     — string GameStatus, per-user game state
game_game_status_<gameId>     — JSON {expectedPlayers, players[]}, ready-tracking
```

## Glossary

| Term | Meaning |
|---|---|
| `op` | Operation identifier string prepended to error messages for tracing |
| `koanf` | Config library used for YAML + env loading |
| `AppError` | Unified error type in `pkg/err_app` with both HTTP and gRPC status codes |
| `ForwardAuth` | Traefik middleware that delegates authentication to Auth Service |
| `ULID` | Sortable unique ID used for `matchId` |
| `Asynq` | Redis-backed async task queue; used for deferred `game:completed` event |
| `gocron` | In-process cron scheduler; used by Match Service every 15 seconds |
| `gobwas/ws` | Low-level WebSocket library (not gorilla/websocket); requires manual frame handling |
| `sarama` | Kafka client library by IBM (formerly Shopify) |
| `lumberjack` | Log file rotation library |
| `ozzo-validation` | Fluent struct validation library |
| `sql-migrate` | SQL migration runner; uses `-- +migrate Up/Down` comments in SQL files |

## Things Future AI Assistants Should Check Before Making Changes

1. **If touching auth flow**: check both the HTTP handler (`delivery/http/handler.go`) and the gRPC handler (`delivery/grpc/handler.go`) — they serve different callers.
2. **If touching Game Service connections**: the `connections` map requires a mutex before any concurrent access fix.
3. **If adding a new Kafka topic**: add the constant to `contract/event/events.go` and handle it in the correct `getTopics()` method.
4. **If changing scoring**: the logic is entirely in `services/game_app/repository/game.go:CalculateScore` and `SavePlayerAnswer`. Config params are in `repository.ScoreConfig`.
5. **If changing question selection**: the active code path is `GetRandomQuestions` in `services/question_app/repository/questions.go`. `GetProperQuestions` exists but is not called.
6. **If deploying to Kubernetes**: check `infra/kubernetes/volume/` for secrets and configmaps that must exist before deployments.
7. **If changing config keys**: update both the YAML file in `infra/deploy/<svc>/development/config.yaml` and the config struct in `services/<svc>_app/config.go`.
8. **Protobuf changes**: after editing `.proto` files, regenerate Go code with `protoc` and commit the generated files.
