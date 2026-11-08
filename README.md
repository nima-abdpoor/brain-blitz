# BrainBlitz

A real-time multiplayer quiz game platform built with Go microservices.

For a prepared single-computer deployment and Telegram connectivity, see
[Run on another computer](docs/run-on-another-computer.md).

## Project Overview

BrainBlitz matches players into live quiz games by category, delivers timed questions over WebSocket connections, scores answers in real time, and presents a final leaderboard when the game ends. The system is composed of five independently deployable Go services communicating over HTTP, gRPC, Kafka, and WebSocket.

## Architecture Overview

```
Client
  │
  ▼
Traefik (API Gateway / Auth Middleware)
  │
  ├── HTTP  ──► Auth Service    (JWT issue & validation)          :5000 HTTP, :6000 gRPC
  ├── HTTP  ──► User Service    (registration, login, profile)    :5001 HTTP, :6001 gRPC
  ├── HTTP  ──► Match Service   (waiting list management)         :5002 HTTP
  ├── WS    ──► Game Service    (real-time game orchestration)    :5003 HTTP/WS
  └── HTTP  ──► Question Svc   (question management)             :5004 HTTP

Kafka topics
  GAME_V1_JOIN_MATCH_QUEUE_REQUESTED  ──► Match Service
  matchMaking_v1_matchUsers           ──► Game Service, Question Service
  question_v1_questions               ──► Game Service

Databases
  PostgreSQL  ── User Service (users table), Question Service (questions, history)
  MongoDB     ── Game Service (game sessions, player answers)  [3-node replica set]
  Redis       ── Match Service (waiting lists), Game Service (game state, task queue)
```

See [docs/architecture.png](docs/architecture.png) for a visual overview.

## Folder Structure

```
brain-blitz/
├── cmd/                        # Service entry points
│   ├── auth/main.go
│   ├── game/main.go
│   ├── match/main.go
│   ├── question/main.go
│   └── user/main.go
├── services/                   # Business logic per service
│   ├── auth_app/
│   ├── game_app/
│   ├── match_app/
│   ├── question_app/
│   └── user_app/
│       ├── app.go              # Application wiring
│       ├── config.go           # Config struct
│       ├── delivery/
│       │   ├── http/           # HTTP handlers, server, routes
│       │   └── grpc/           # gRPC handlers, server
│       ├── repository/         # Data access layer
│       │   └── migrations/     # SQL migration files
│       └── service/            # Core business logic, entities, params
├── adapter/                    # Infrastructure adapters (Redis, Kafka, WebSocket, Asynq, Auth gRPC)
├── contract/                   # Protobuf definitions and generated Go code
│   ├── auth/
│   ├── match/
│   └── question/
├── pkg/                        # Shared libraries
│   ├── cache_manager/
│   ├── cfg_loader/
│   ├── common/
│   ├── email/
│   ├── err_app/
│   ├── err_msg/
│   ├── grpc/
│   ├── http_server/
│   ├── json/
│   ├── logger/
│   ├── mongo/
│   ├── postgresql/
│   └── postgresqlmigrator/
├── infra/
│   ├── deploy/                 # Per-service Dockerfiles and config.yaml files
│   │   ├── auth/development/
│   │   ├── game/development/
│   │   ├── match/development/
│   │   ├── question/development/
│   │   └── user/development/
│   └── kubernetes/             # K8s deployment, service, volume, and ingress manifests
├── docs/                       # Documentation
├── .github/workflows/          # CI/CD pipelines
├── docker-compose.yml
├── prometheus.yml
├── buf.yaml / buf.gen.yaml     # Protobuf toolchain config
├── go.mod
└── Makefile
```

## Installation

### Prerequisites

| Tool | Version |
|---|---|
| Go | 1.23+ |
| Docker | 20+ |
| Docker Compose | 1.29+ |
| `protoc` + `buf` | For regenerating Protobuf (optional) |

### Clone

```bash
git clone https://github.com/nima-abdpoor/brain-blitz.git
cd brain-blitz
```

## Configuration

Each service reads configuration from two sources, with environment variables taking precedence:

1. A YAML file at `infra/deploy/<service>/development/config.yaml`
2. Environment variables with a per-service prefix (`AUTH_`, `USER_`, `GAME_`, `MATCH_`, `QUESTION_`)

Key separator for nested config keys is `__` (double underscore). Example:

```bash
# Override the Postgres host for the User service:
USER_postgres_db__host=db-prod-host
```

### Service Ports

| Service | HTTP | gRPC |
|---|---|---|
| Auth | 5000 | 6000 |
| User | 5001 | 6001 |
| Match | 5002 | — |
| Game | 5003 | — |
| Question | 5004 | — |

### Auth Service (`infra/deploy/auth/development/config.yaml`)

```yaml
service:
  secret_key: SECRET_KEY          # Override with AUTH_service__secret_key env var
  access_token_expire_time: 24h
  refresh_token_expire_time: 120h
```

### Game Service (`infra/deploy/game/development/config.yaml`)

```yaml
repository:
  valid_answer_timeout: 2m2s
  score:
    base_score: 5
    max_bonus: 10
    bonus_deadline: 115s
```

### Match Service (`infra/deploy/match/development/config.yaml`)

```yaml
scheduler:
  interval: 15          # seconds between matchmaking runs
service:
  waiting_timeout: "20m"
```

## Running Locally

Docker Compose v2 and Python 3 are required; Go runs in the build container.
Before starting Redis, follow [Redis host setup](docs/redis-host-setup.md) to
check and configure `vm.overcommit_memory = 1` on the Linux kernel running Docker.
From the project directory:

```sh
python3 scripts/init-env.py
docker compose config --quiet
docker compose --parallel 2 build
docker compose up -d --wait --wait-timeout 300
docker compose --profile tools run --rm seed-demo
curl --fail http://127.0.0.1:8088/user-service/public/api/v1/health-check
```

This starts all five application services, Kafka, PostgreSQL, Redis, a three-node
MongoDB replica set, and Traefik. Initializer jobs create the Question database,
Kafka topics, and replica set before their dependent applications start. The
optional demo import adds 30 questions and can be repeated safely.

The gateway binds to `127.0.0.1:8088` by default. Database and administrative
ports are private. Set `HTTP_BIND` and `HTTP_PORT` in `.env` when appropriate for
your deployment. The generated `.env` contains machine-local secrets and must
not be committed. Monitoring is not part of the default single-computer stack.

See [the deployment guide](docs/run-on-another-computer.md) for resource
requirements, bot connectivity, an end-to-end smoke test, and data backups.

## Docker

Each service has a multi-stage Dockerfile at `infra/deploy/<service>/development/Dockerfile`.

Build a specific service manually:

```bash
docker build -f infra/deploy/auth/development/Dockerfile \
  --build-arg GO_IMAGE_NAME=golang \
  --build-arg GO_IMAGE_VERSION=1.26 \
  -t brain-blitz/auth .
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...
```

Unit tests exist for:
- `services/auth_app/service/` — token creation and validation logic
- `services/match_app/service/` — matchmaking and validation logic
- `services/user_app/service/` — user service logic

Integration tests exist for:
- `services/auth_app/delivery/http/validate_tokne_integration_test.go`
- `services/match_app/delivery/http/add_to_waiting_list_integration_test.go`
- `services/user_app/delivery/http/login_integration_test.go`
- `services/user_app/repository/user_test.go`

## Linting and Formatting

```bash
go vet ./...
gofmt -w .
```

## Migrations

Migrations are embedded as SQL files and run automatically on service startup using `pkg/postgresqlmigrator`.

| Service | Migration directory |
|---|---|
| User | `services/user_app/repository/migrations/` |
| Question | `services/question_app/repository/migrations/` |

To install the standalone `sql-migrate` CLI:

```bash
make install-sql-migrate
```

## Deployment

### Docker Compose (Single Computer)

Follow [Run on another computer](docs/run-on-another-computer.md).

### Kubernetes

See [docs/kubernetes-init.md](docs/kubernetes-init.md) for full Kubernetes setup.

Manifests are in `infra/kubernetes/`:
- `deployment/` — Deployment and Service objects for each microservice and infrastructure component
- `volume/` — ConfigMaps, Secrets, PersistentVolumeClaims
- `ingress/` — Traefik IngressRoute definitions

```bash
kubectl apply -f infra/kubernetes/volume/
kubectl apply -f infra/kubernetes/deployment/
kubectl apply -f infra/kubernetes/ingress/
```

## API Overview

All protected endpoints require `Authorization: Bearer <accessToken>`, validated by Traefik's ForwardAuth middleware calling Auth Service. Raw access tokens remain accepted for compatibility. Token minting endpoints are internal and are not exposed by the default Compose gateway.

### Auth Service (`/auth-service`)

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/access-token` | No | Create JWT access token |
| POST | `/api/v1/refresh-token` | No | Create JWT refresh token |
| GET/POST | `/api/v1/validate-token` | No | Validate JWT (used by Traefik) |
| GET | `/api/v1/health-check` | No | Health check |

### User Service (`/user-service`)

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/public/api/v1/signup` | No | Register new user |
| POST | `/public/api/v1/login` | No | Login, returns access and refresh tokens |
| GET | `/api/v1/profile` | Yes | Get current user profile |

### Match Service (`/match-service`)

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/addToWaitingList` | Yes | Add user to matchmaking queue |
| GET | `/api/v1/health-check` | No | Health check |

### Game Service (`/game-service`)

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/v1/process-game` | Yes | Upgrade to WebSocket for game session |
| GET | `/api/v1/health-check` | No | Health check |

The Game Service WebSocket endpoint handles a JSON command protocol. See [docs/context/03-business-domain.md](docs/context/03-business-domain.md) for the full WebSocket protocol.

### Question Service

Internal service only. No public routes beyond health check.

## Monitoring

The single-computer Compose stack omits Prometheus, Grafana, and the MongoDB exporter to reduce resource use. Monitoring configuration remains in `prometheus.yml` for a separate monitoring deployment. Administrative monitoring endpoints are not exposed by the default stack.

> Note: Application-level metrics (service latencies, error rates) are not yet instrumented. Many places in the code contain `// todo add metrics`.

## Replication Setup

- [MongoDB Replication Guide](docs/mongodb-replication.md)
- [PostgreSQL Replication Guide](docs/postgresql-replication.md)
- [PostgreSQL vs. MongoDB Replication Comparison](docs/PostgreSQL-vs-MongoDB-replication.md)

## Development Workflow

### Branch Strategy

The default branch is `develop`. All PRs target `develop`.

On every push to `develop`:
- GitHub Actions builds and tests all packages
- A new patch-version git tag is created automatically
- Docker images are built (but not pushed) for services: auth, match, game, user

### Protobuf

Proto files live in `contract/<service>/proto/`. Generated Go code is committed alongside.

To regenerate Go files from a `.proto` file:

```bash
protoc --go_out=. --go-grpc_out=. contract/<service>/proto/<file>.proto
```

Ensure the `go_package` option matches the target directory:

```proto
option go_package = "contract/<service>/golang";
```

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| Services cannot connect to Kafka | Kafka not ready yet | Wait for Kafka health check to pass; services retry 5× with 2s delay |
| MongoDB writes fail | Replica set not initialized | Run `bb-mongo-init` container or manually call `rs.initiate()` |
| `AUTH_service__secret_key` not set | JWT signing will use `SECRET_KEY` default | Set the env var in docker-compose or K8s secret |
| `X-User-ID` header missing | Traefik ForwardAuth not configured, or request sent to public route | Ensure protected routes go through `auth` middleware |
| Matchmaking never fires | Match scheduler interval | Default is every 15 seconds; at least 2 players in same category required |

## Contribution Guide

1. Fork the repository and create a feature branch from `develop`.
2. Follow the existing package structure: `delivery/` → `service/` → `repository/`.
3. Add or update unit tests for any business logic change.
4. If adding a new Kafka topic, add the topic constant to `contract/event/events.go`.
5. If changing a Protobuf contract, regenerate the Go code and commit it.
6. Open a PR targeting `develop` — CI must pass before merge.
