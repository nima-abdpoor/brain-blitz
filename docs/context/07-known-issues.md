# 07 — Known Issues

## Bugs

سیفثقتنغعغفققم۴صقثظ### 1. ~~Auth adapter `GetRefreshToken` returns an access token~~ **FIXED** (Phase 1 / R-01)

**File**: `adapter/auth/client.go`

`GetRefreshToken` was calling `GetAccessToken` on the gRPC stub instead of `GetRefreshToken`. Fixed: now builds `CreateRefreshTokenRequest`, calls `GetRefreshToken`, and returns `res.RefreshToken`. Test: `adapter/auth/client_test.go`.

### 2. ~~`fmt.Println` leaks JWT secret key to stdout~~ **FIXED** (Phase 1 / R-02)

**File**: `services/auth_app/service/service.go`

The debug print `fmt.Println(svc.config.SecretKey, signedString, err)` was removed. Existing `TestCreateAccessToken` test continues to pass.

### 3. ~~In-memory WebSocket connection map is not thread-safe~~ **FIXED** (Phase 1 / R-03)

**File**: `services/game_app/service/service.go`

`svc.connections` is a `map[uint64]net.Conn` modified from multiple goroutines. Fixed: `*sync.RWMutex` field added to `Service`; all map reads use `RLock/RUnlock`, all writes use `Lock/Unlock`. Test: `services/game_app/service/service_test.go` (run with `-race`).

### 4. ZRange parameters are reversed in `GetWaitingListByCategory`

**File**: `services/match_app/repository/match.go:43`

```go
mintTime := int(time.Now().Add(m.Config.MinTimeWaitingListSelection).UnixMicro())
maxTime := int(time.Now().UnixMicro())
```

`MinTimeWaitingListSelection` is configured as `-20m`, so `mintTime` is 20 minutes in the past. However the `ZRange` call passes `(mintTime, maxTime)` — start and stop for a time-ordered sorted set. The Redis `ZRangeByScore` command receives `Min: mintTime, Max: maxTime` which is the correct range (past → now). This is fine, but the variable names are confusing.

### 5. Match Service config includes `GRPCServer` but no gRPC server is started

**File**: `services/match_app/config.go:9`, `services/match_app/app.go`

`Config` includes `GRPCServer grpc.Config` but `app.go` never creates or starts a gRPC server. This dead config field is misleading.

### 6. ~~`UpsertReadyPlayer` ready-check logic returns `true` prematurely~~ **FIXED** (Phase 1 / R-04)

**File**: `services/game_app/repository/game.go`

The `== 1` branch in the early-return guard was removed. The final return now computes `len(gs.Players) == gs.ExpectedNumberOfPlayers` so the game starts only after every player's READY state is persisted. Test: `services/game_app/repository/game_test.go`.

## Technical Debt

### 7. Metrics not implemented

Over 15 locations in the codebase contain `// todo add metrics`. No application-level Prometheus metrics are emitted by any service. Only MongoDB is monitored (via exporter).

### 8. Kafka consumer does not use consumer groups

**File**: `adapter/broker/kafka-broker.go:79`

The `Consume` function uses `consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)` — a low-level partition consumer that:
- Only reads partition 0
- Does not use consumer groups (no offset tracking, no rebalancing)
- Will not scale horizontally (multiple instances would each independently consume the same messages)

### 9. Connection retry not implemented for PostgreSQL

**File**: `cmd/user/main.go:42`, `cmd/question/main.go:41`

```go
// todo retry to connect in result of connection failure
postgresConn, cnErr := postgresql.Connect(cfg.PostgresDB)
```

If PostgreSQL is temporarily unavailable at startup, the service dies. No retry logic.

### 10. Waiting list users not removed after match in all paths

**File**: `services/match_app/service/service.go:133`

```go
// todo remove these users from waiting list
if len(finalUsers) > 0 {
```

The comment indicates awareness that removal is needed. Removal is actually implemented in `publishFinalUsers()`, but the comment is a leftover from a partial implementation.

### 11. `GetProperQuestions` (history-aware) is never called

**File**: `services/question_app/service/service.go:53`

`GetRandomQuestions` is always used. The history-aware `GetProperQuestions` (which excludes questions seen in the past 30 days) exists in the repository but is never called by the service. The `user_question_history` table is therefore never populated and question deduplication does not work.

### 12. `question_app` users table conflicts with `user_app` users table

**File**: `services/question_app/repository/migrations/1746384268-create_users_table.sql`

The Question Service migration creates a `users` table with `UUID` primary key, while User Service has a `users` table with `SERIAL` primary key. These are in different databases (`question_db` vs `user_db`) so there is no actual conflict, but the naming is confusing and the Question Service `users` table is never populated.

### 13. `AddQuestion` endpoint is a stub

**File**: `services/question_app/service/service.go:32`

```go
func (svc Service) AddQuestion(...) (AddQuestionResponse, error) {
    return AddQuestionResponse{}, nil
}
```

No implementation. There is no way to add questions to the database via the API.

### 14. Question Service not in Kubernetes manifests

`infra/kubernetes/deployment/` contains manifests for `auth`, `game`, `match`, `user` but not `question`. The Question Service cannot be deployed to Kubernetes without adding this.

### 15. Game completion does not update Game document status

After the game completes (Asynq task fires), the game's `status` field in MongoDB is never updated to `FINISHED`. The document remains with status `PENDING` indefinitely.

### 16. Makefile references non-existent path

**File**: `Makefile:41`

```makefile
sqlc-generate:
    sudo sqlc generate --file internal/infra/repository/sqlc/sqlc.yml
```

The path `internal/infra/repository/sqlc/sqlc.yml` does not exist. `sqlc` was likely considered but not adopted; this target is dead.

### 17. `wg.Add(3)` in Game Service `startServers` is placed before goroutines

**File**: `services/game_app/app.go:97`

`wg.Add(3)` is called outside the goroutines but the goroutines run `defer wg.Done()`. If any goroutine panics before `defer` is registered, the WaitGroup count would be wrong. This is a minor issue as `defer` is the first statement, but the pattern is fragile.

## TODOs from Comments

| File | TODO |
|---|---|
| `services/auth_app/app.go:74,85` | `// todo add metrics` |
| `services/auth_app/service/service.go:124` | `// todo add metrics` |
| `services/user_app/service/service.go:84` | `// todo add to metrics` |
| `services/user_app/delivery/http/handler.go` (Profile) | `// todo check if logger needed`, `// todo add metrics` |
| `cmd/user/main.go:41` | `//todo retry to connect in result of connection failure` |
| `cmd/question/main.go:40` | `//todo retry to connect in result of connection failure` |
| `services/game_app/service/service.go:121` | Commented-out status reset logic |
| `services/game_app/repository/game.go:46` | `//todo check the possibility of conversion` |
| `services/match_app/service/service.go:133` | `// todo remove these users from waiting list` |
| `services/user_app/repository/user.go:45` | `// should we do this automatically by defining function in postgres or not?` |

## Security Concerns

1. **JWT secret key logged to stdout** (see Bug #2 above) — critical.
2. **JWT secret default value** is `SECRET_KEY` (literal string) — insecure default.
3. **MongoDB exporter** is exposed on port `9216` with `--collect-all` — sensitive metrics are publicly accessible without authentication in the default docker-compose setup.
4. **Redis** runs with `--protected-mode no` and no password in docker-compose.
5. **Traefik dashboard** (`localhost:8080`) is in `--api.insecure=true` mode.
6. **CORS** is configured as `allow_origins: "*"` across all services.
