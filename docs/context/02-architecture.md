# 02 — Architecture

## Package Responsibilities

### `cmd/`

Entry points only. Each `main.go`:
1. Resolves the working directory
2. Loads config via `cfgloader.Load` (YAML + env vars)
3. Initialises the global logger
4. Opens database/infrastructure connections
5. Calls `<service>.Setup(cfg, ...)` to wire dependencies
6. Calls `app.Start()` to block until shutdown

### `services/<name>_app/`

| Sub-package | Responsibility |
|---|---|
| `app.go` | Wires all dependencies; starts/stops HTTP, gRPC, Kafka consumer, scheduler goroutines |
| `config.go` | Flat config struct aggregating configs from pkg and adapter layers |
| `delivery/http/` | Echo handlers + route registration; translates HTTP to service calls |
| `delivery/grpc/` | gRPC handlers; translates proto types to service types |
| `repository/` | Database access; implements the `Repository` interface defined in `service/` |
| `service/` | Business logic; defines `Repository` interface, entity types, param types, validators |

### `adapter/`

Infrastructure wrappers that are not tied to a single service:

| Package | Wraps |
|---|---|
| `adapter/auth` | gRPC client for the Auth Service (`TokenClient` interface) |
| `adapter/broker` | Kafka producer/consumer via `sarama` (`Broker` interface) |
| `adapter/redis` | Redis client; exposes sorted set, get/set operations |
| `adapter/task-queue` | Asynq publisher and worker; wraps task enqueueing and handler registration |
| `adapter/websocket` | WebSocket upgrade, read, write primitives via `gobwas/ws` |

### `contract/`

Protobuf definitions and generated Go stubs:

| Directory | Content |
|---|---|
| `contract/auth/proto/` | Auth service RPC contract |
| `contract/auth/golang/` | Generated Go code |
| `contract/match/proto/` | Match event messages |
| `contract/match/golang/` | Generated Go code |
| `contract/question/proto/` | Question event messages |
| `contract/question/golang/` | Generated Go code |
| `contract/event/events.go` | Topic name constants |

### `pkg/`

Shared utilities used across services:

| Package | Purpose |
|---|---|
| `pkg/cfg_loader` | Loads config from YAML file + env vars via `koanf` |
| `pkg/logger` | Global `slog`-based structured JSON logger with file rotation (lumberjack) |
| `pkg/http_server` | Echo wrapper: creates router with CORS middleware, start/stop methods |
| `pkg/grpc` | gRPC server and client factory functions |
| `pkg/err_app` | Unified `AppError` type with HTTP + gRPC status codes; wrapping and conversion utilities |
| `pkg/err_msg` | Human-readable error message string constants |
| `pkg/cache_manager` | Redis-based cache wrapper (get, set, delete, TTL) |
| `pkg/postgresql` | PostgreSQL connection factory (`database/sql` + `lib/pq`) |
| `pkg/mongo` | MongoDB driver factory; supports replica sets |
| `pkg/postgresqlmigrator` | Runs SQL migrations using `rubenv/sql-migrate` |
| `pkg/common` | bcrypt helpers (hash/check password), UTC millisecond timestamp |
| `pkg/email` | Email validation |
| `pkg/json` | JSON helpers |

## Dependency Graph

```
cmd/auth        → services/auth_app    → pkg/grpc, pkg/http_server, pkg/logger, pkg/err_app
cmd/user        → services/user_app    → adapter/auth, adapter/redis, pkg/postgresql, pkg/grpc
cmd/match       → services/match_app   → adapter/broker, adapter/redis, pkg/http_server
cmd/game        → services/game_app    → adapter/broker, adapter/redis, adapter/websocket,
                                         adapter/task-queue, pkg/mongo
cmd/question    → services/question_app→ adapter/broker, pkg/postgresql

services/*_app  → contract/            (protobuf types for Kafka messages)
adapter/auth    → contract/auth/golang (gRPC stubs)
```

## Request Lifecycle

### Public HTTP request (e.g. POST /user-service/public/api/v1/signup)

```
Client
  → Traefik (strip /user-service prefix, no auth middleware)
  → User Service :5001
  → delivery/http/Handler.SignUp
  → service.Service.SignUp
  → repository.UserRepository.InsertUser (PostgreSQL)
  ← response JSON
```

### Protected HTTP request (e.g. GET /user-service/api/v1/profile)

```
Client (Authorization: Bearer <token>)
  → Traefik
  → ForwardAuth: POST Auth Service :5000/api/v1/validate-token
    → auth service decodes JWT, returns X-User-ID, X-User-Role, X-Auth-Data headers
  → User Service :5001
  → delivery/http/Handler.Profile reads X-User-ID header
  → service.Service.Profile → PostgreSQL lookup
  ← response JSON
```

### WebSocket game session (GET /game-service/api/v1/process-game)

```
Client (Authorization: Bearer <token>)
  → Traefik ForwardAuth → Auth Service validates token
  → Game Service :5003, X-User-ID header injected
  → delivery/http/Handler.ProcessGame
  → service.Service.ProcessGame
    → websocket.Upgrade (raw TCP connection hijacked)
    → goroutine: read loop processes JSON command frames
    ← JSON event frames pushed to client
```

### Kafka event flow

```
Game Service
  (on ADD_TO_WAITING_LIST command from WebSocket client)
  → broker.Publish(GAME_V1_JOIN_MATCH_QUEUE_REQUESTED, proto-encoded AddToWaitingList)

Match Service
  (consumer goroutine)
  ← broker.Consume(GAME_V1_JOIN_MATCH_QUEUE_REQUESTED)
  → repository.AddToWaitingList (Redis ZAdd)

Match Scheduler (every 15s)
  → service.MatchWaitUsers
  → repository.GetWaitingListByCategory (Redis ZRange)
  → pairs players
  → broker.Publish(matchMaking_v1_matchUsers, proto-encoded AllMatchedUsers)
  → repository.RemoveWaitingMember (Redis ZRem)

Question Service                          Game Service
  ← Consume(matchMaking_v1_matchUsers)     ← Consume(matchMaking_v1_matchUsers)
  → GetRandomQuestions (PostgreSQL)         → CreateGame (MongoDB)
  → Publish(question_v1_questions)          → notify players via WebSocket (MATCH_CREATED)
                                            → saveUsersGameStatus → Redis

Game Service
  ← Consume(question_v1_questions)
  → SaveQuestionsByMatchId (MongoDB + Redis)

Client sends READY command
  → Game Service sets answer deadlines for all questions (Redis)
  → sends all questions to all players (WebSocket)
  → enqueues Asynq task "game:completed" with delay = total answer time

Client sends ANSWER command
  → Game Service validates answer time window
  → saves PlayerAnswer to MongoDB
  → returns current leaderboard

Asynq worker fires "game:completed"
  → GetLeaderBoard (MongoDB aggregate)
  → sends final leaderboard to all players via WebSocket
  → closes WebSocket connections
```

## Startup Flow

Each service startup follows the same pattern:

```
main()
  cfgloader.Load → Config struct
  logger.Init    → global slog logger
  Connect databases (PostgreSQL / MongoDB)
  Run migrations (user, question services)
  grpc.NewClient (user service connects to auth)
  <service>.Setup(cfg, ...) → Application struct
  app.Start()
    → signal.NotifyContext(SIGINT, SIGTERM)
    → startServers() → goroutines for HTTP, gRPC, Kafka consumer, scheduler
    → block on ctx.Done()
    → shutdownServers(timeout)
    → wg.Wait()
```

## Important Interfaces

### `service.Repository` (per service)

Each service defines its own `Repository` interface in `service/service.go`. The concrete implementation lives in `repository/`. This inversion keeps business logic testable without a real database.

### `broker.Broker` (`adapter/broker/Broker.go`)

```go
type Broker interface {
    Publish(ctx context.Context, topic string, message []byte) error
    Consume(ctx context.Context, topic string, handler func([]byte, context.Context) error) error
}
```

### `websocket.WebSocket` (`adapter/websocket/websocket.go`)

```go
type WebSocket interface {
    Upgrade(r *http.Request, w http.ResponseWriter) (*net.Conn, *bufio.ReadWriter, Handshake, error)
    ReadClientData(rw io.ReadWriter) (string, OpCode, error)
    WriteServerData(rw io.Writer, code OpCode, message string) error
}
```

### `auth_adapter.TokenClient` (`adapter/auth/client.go`)

```go
type TokenClient interface {
    GetAccessToken(ctx context.Context, req CreateAccessTokenRequest) (CreateAccessTokenResponse, error)
    GetRefreshToken(ctx context.Context, req CreateRefreshTokenRequest) (CreateRefreshTokenResponse, error)
}
```

### `taskqueue.TaskPublisher` / `taskqueue.TaskProcessor` (`adapter/task-queue/task-queue-manager.go`)

Used by Game Service for deferred "game:completed" task scheduling.

## Concurrency Model

- Each service starts goroutines for its servers at startup; `sync.WaitGroup` tracks them.
- Kafka consumers run one goroutine per topic per service.
- Match Service scheduler runs in its own goroutine; receives stop signal via a `chan bool`.
- Game Service uses a shared in-process `map[uint64]net.Conn` (`IdToConnection`) for active WebSocket connections. This is **not thread-safe under concurrent writes** and limits the Game Service to a single process instance.
- Asynq worker (Game Service) runs its own pool with concurrency=10.

## Error Handling

All business logic returns `*errApp.AppError` which carries:
- `OP` — operation name for tracing
- `Code` — machine-readable code (e.g. `"NOT_FOUND"`)
- `Message` — human-readable message
- `HTTPStatus` — HTTP status code
- `GRPCStatus` — gRPC status code

Delivery layer calls `errApp.ToHTTPJson(err)` or `errApp.ToGRPCJson(err)` to serialize.

Pre-defined errors in `pkg/err_app/errors.go`:
- `ErrNotFound` — 404 / NOT_FOUND
- `ErrInternal` — 500 / INTERNAL
- `ErrInvalidInput` — 400 / INVALID_ARGUMENT
- `ErrUnauthorized` — 401 / UNAUTHENTICATED
- `ErrInvalidLOGIN` — 403 / PERMISSION_DENIED

## Logging

Global singleton `slog.Logger` initialized in `pkg/logger`. Output: JSON to stdout AND a rotating log file (via lumberjack). Each service writes to its own log file path (e.g. `logs/auth/service.log`).

Log calls follow the key-value pattern: `logger.Error("message", "key", value, ...)`.

## Configuration Loading

`pkg/cfg_loader.Load(Option, &config)` uses `koanf`:
1. Loads YAML file (base defaults)
2. Loads environment variables with the given prefix (e.g. `GAME_`)
3. Transforms env var keys: strip prefix, lowercase, replace `__` with `.`

Example: `GAME_mongo__host=bb-mongo1` → `mongo.host = "bb-mongo1"`
