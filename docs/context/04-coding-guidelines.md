# 04 — Coding Guidelines

## Code Style

- Go 1.23. Follow standard `gofmt` formatting.
- Use `go vet ./...` for static analysis before committing.
- All error values should be of type `*errApp.AppError` at service boundaries; use `errApp.Wrap` or `errApp.New`.

## Package Organization

Each service follows a consistent three-layer structure:

```
<service>_app/
├── app.go          ← composition root; no business logic
├── config.go       ← aggregate config struct
├── delivery/
│   ├── http/
│   │   ├── handler.go       ← HTTP handlers (bind, call service, return JSON)
│   │   ├── server.go        ← route registration
│   │   └── health-check.go  ← health endpoint
│   └── grpc/
│       ├── handler.go       ← gRPC handlers (map proto ↔ service types)
│       └── server.go        ← gRPC server setup
├── repository/
│   ├── <name>.go            ← implements service.Repository interface
│   ├── helper.go            ← internal repository helpers
│   ├── param.go             ← repository-specific structs (if needed)
│   └── migrations/          ← SQL files
└── service/
    ├── service.go   ← Repository interface + Service struct + business methods
    ├── entity.go    ← domain types, enums, mapping functions
    ├── param.go     ← request/response types
    ├── helper.go    ← pure functions used by service
    ├── validator.go ← input validation using go-ozzo
    └── <name>_test.go
```

## Naming Conventions

| Item | Convention | Example |
|---|---|---|
| Packages | `snake_case`, short | `auth_app`, `err_app`, `task-queue` |
| Types | `PascalCase` | `Service`, `GameStatus`, `AppError` |
| Interfaces | `PascalCase`, noun | `Repository`, `Broker`, `WebSocket` |
| Methods | `PascalCase` (exported), `camelCase` (unexported) | `CreateGame`, `saveUsersGameStatus` |
| Config fields | match koanf tag | `koanf:"http_server"` → struct field `HTTPServer` |
| Constants | `PascalCase` for domain, `SCREAMING_SNAKE` for Kafka topics | `GameStatusPending`, `GAME_V1_JOIN_MATCH_QUEUE_REQUESTED` |
| Error variables | `Err` prefix | `ErrNotFound`, `ErrInternal` |

## Patterns Used

### Dependency Injection via Constructor

```go
func NewService(repo Repository, cm CacheManager, grpc TokenClient, logger Logger) Service {
    return Service{repository: repo, CacheManager: cm, grpcClient: grpc, Logger: logger}
}
```

All dependencies are explicit constructor parameters. No globals (except the logger singleton).

### Interface-based Abstraction

Business logic in `service/` depends only on interfaces (`Repository`, `Broker`, `WebSocket`). This enables unit testing with mock implementations without touching infrastructure code.

### Operation Tracking

Every function that can fail defines a local `op` constant:
```go
const op = "service.SignUp"
return errApp.Wrap(op, err, errApp.ErrInternal, data, s.Logger)
```

This propagates the call chain context into error logs.

### Entity Mapping at Boundaries

Proto types are never used inside `service/`. Conversion happens in the delivery layer or at consumer/producer call sites:
```go
// In grpc/handler.go
for _, data := range req.GetData() {
    requestData = append(requestData, service.CreateTokenRequest{Key: data.GetKey(), Value: data.GetValue()})
}
```

### Graceful Shutdown Pattern

Every service `app.go` implements the same pattern:
1. `signal.NotifyContext` for OS signals
2. `startServers()` launches goroutines with `sync.WaitGroup`
3. Block on `ctx.Done()`
4. `shutdownServers(timeoutCtx)` calls stop methods concurrently
5. `wg.Wait()` to drain all goroutines

## Validation

Input validation uses `github.com/go-ozzo/ozzo-validation/v4`. Validators live in `service/validator.go`. Example:

```go
func ValidateCreateAccessTokenRequest(req CreateAccessTokenRequest) error {
    return validation.ValidateStruct(&req,
        validation.Field(&req.Data, validation.Required.Error("data is required"), ...),
    )
}
```

## Testing Strategy

| Layer | Approach |
|---|---|
| `service/` | Unit tests with mock `Repository` (interface), mock adapters |
| `repository/` | Integration tests using `go-sqlmock` or real DB |
| `delivery/http/` | Integration tests against a real running service instance |

Tests use `github.com/stretchr/testify` (`assert` / `require` / `mock`).

Test file naming: `<file>_test.go` alongside the file under test.

## Anti-Patterns to Avoid

1. **Do not** put infrastructure code (SQL, Redis, Kafka) inside `service/`. Use the `Repository` or `Broker` interface.
2. **Do not** import proto-generated types inside `service/`. Map at boundaries.
3. **Do not** use `fmt.Println` for observability — use the structured logger.
4. **Do not** hard-code magic strings for Kafka topics or Redis key prefixes — use constants in `contract/event/events.go` or configuration.
5. **Do not** share a single Redis sorted set across categories — each category has its own key.
6. **Do not** skip graceful shutdown — always handle `SIGTERM`.

## Config Keys Convention

Config is loaded with:
- YAML key: `snake_case` nested (e.g. `http_server.port`)
- Env var: `<PREFIX>_<section>__<key>` (single underscore between prefix and section, double underscore as level separator)

Example for Game service:
- YAML: `mongo.host: bb-mongo1`
- Env: `GAME_mongo__host=bb-mongo1`
