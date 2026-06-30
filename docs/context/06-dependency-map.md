# 06 — Dependency Map

## Internal Package Relationships

```
cmd/auth
  └─► services/auth_app
        └─► pkg/grpc, pkg/http_server, pkg/logger, pkg/err_app, pkg/err_msg
            pkg/cfg_loader
            contract/auth/golang  (gRPC server stubs)

cmd/user
  └─► services/user_app
        ├─► adapter/auth         (gRPC client to Auth Service)
        ├─► adapter/redis        (optional cache)
        ├─► pkg/cache_manager
        ├─► pkg/postgresql
        ├─► pkg/postgresqlmigrator
        ├─► pkg/grpc
        ├─► pkg/http_server
        ├─► pkg/logger, pkg/err_app, pkg/err_msg, pkg/common, pkg/email
        └─► contract/auth/golang (via adapter/auth)

cmd/match
  └─► services/match_app
        ├─► adapter/broker       (Kafka)
        ├─► adapter/redis        (waiting lists)
        ├─► pkg/http_server, pkg/logger, pkg/err_app, pkg/err_msg
        └─► contract/match/golang, contract/event

cmd/game
  └─► services/game_app
        ├─► adapter/broker       (Kafka consumer + producer)
        ├─► adapter/redis        (game state)
        ├─► adapter/websocket    (raw WebSocket)
        ├─► adapter/task-queue   (Asynq publisher + worker)
        ├─► pkg/mongo
        ├─► pkg/http_server, pkg/logger, pkg/err_app, pkg/err_msg
        └─► contract/match/golang, contract/question/golang, contract/event

cmd/question
  └─► services/question_app
        ├─► adapter/broker       (Kafka consumer + producer)
        ├─► pkg/postgresql
        ├─► pkg/postgresqlmigrator
        ├─► pkg/http_server, pkg/logger, pkg/err_app, pkg/err_msg
        └─► contract/match/golang, contract/question/golang
```

## External Services

| Service | Used by | How |
|---|---|---|
| **Kafka** (Bitnami, KRaft mode) | Match, Game, Question | `adapter/broker.KafkaBroker` via `sarama` |
| **PostgreSQL 17** | User Service, Question Service | `database/sql` + `lib/pq` driver |
| **MongoDB 8.0** (3-node replica set rs0) | Game Service | `go.mongodb.org/mongo-driver` |
| **Redis 6.2** | Match Service (sorted sets), Game Service (game state + Asynq), User Service (optional cache) | `redis/go-redis/v9` |
| **Traefik v2.10** | All public-facing routes | Docker labels / K8s IngressRoute |
| **Prometheus** | Monitoring (MongoDB exporter only) | Scrape at `:9216` |
| **Grafana** | Dashboarding | Reads from Prometheus |
| **Asynq** | Game Service (deferred "game:completed" task) | Uses Redis as backend |

## External Libraries

| Library | Purpose | Services |
|---|---|---|
| `github.com/IBM/sarama` | Kafka client | Match, Game, Question |
| `github.com/labstack/echo/v4` | HTTP framework | All |
| `google.golang.org/grpc` | gRPC framework | Auth, User |
| `google.golang.org/protobuf` | Protobuf serialization | All (Kafka messages) |
| `go.mongodb.org/mongo-driver` | MongoDB client | Game |
| `github.com/redis/go-redis/v9` | Redis client | Match, Game, User |
| `github.com/golang-jwt/jwt/v5` | JWT creation/parsing | Auth |
| `github.com/knadh/koanf/v2` | Config loading (YAML + env) | All |
| `github.com/rubenv/sql-migrate` | SQL migrations | User, Question |
| `github.com/lib/pq` | PostgreSQL driver | User, Question |
| `github.com/gobwas/ws` | Low-level WebSocket | Game |
| `github.com/hibiken/asynq` | Async task queue (Redis-backed) | Game |
| `github.com/go-co-op/gocron` | Cron-style scheduler | Match |
| `github.com/google/uuid` | UUID generation | Auth (JWT ID) |
| `github.com/oklog/ulid/v2` | ULID generation | Match (match IDs) |
| `github.com/go-ozzo/ozzo-validation/v4` | Struct validation | Auth, Match |
| `github.com/thoas/go-funk` | Functional utilities (IndexOf) | Match |
| `golang.org/x/crypto` | bcrypt | User (password hashing) |
| `gopkg.in/natefinch/lumberjack.v2` | Log file rotation | All (via logger pkg) |
| `github.com/stretchr/testify` | Test assertions | Test files |
| `github.com/DATA-DOG/go-sqlmock` | SQL mock for tests | User, Question tests |

## Critical Dependencies

### Kafka startup order

The Match, Game, and Question services all connect to Kafka at startup with 5 retry attempts × 2 second delay. If Kafka is not ready within ~10 seconds, the service panics. In `docker-compose.yml` the Kafka health check must pass first.

### MongoDB replica set

Game Service requires MongoDB to be running as a replica set (`rs0`). The `bb-mongo-init` container initializes the replica set. If Game Service starts before the replica set is ready, MongoDB write operations will fail silently or return errors.

### Auth Service dependency

User Service opens a gRPC connection to Auth Service at startup (`cmd/user/main.go`). If Auth Service is not reachable, User Service will fail to start (gRPC dial fails).

### Redis dependency

Match Service, Game Service, and the Asynq worker all fail without Redis. Redis connectivity is not retried — the service will start but operations will fail at runtime.

## Service Port Summary

| Service | HTTP | gRPC |
|---|---|---|
| Auth | 5000 | 6000 |
| User | 5001 | 6001 |
| Match | 5002 | — |
| Game | 5003 | — |
| Question | 5004 | — |
| Traefik (proxy) | 80 | — |
| Traefik (dashboard) | 8080 | — |
| Prometheus | 9090 | — |
| Grafana | 3000 | — |
| MongoDB exporter | 9216 | — |
| PostgreSQL | 5433 (host) | — |
| MongoDB | 27018 (host) | — |
| Redis | 6379 | — |
| Kafka | 9092 | — |
