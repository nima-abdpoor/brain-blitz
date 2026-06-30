# 05 — Development Context

## Local Setup

### Prerequisites

```bash
# Verify Go 1.23+
go version

# Verify Docker
docker --version
docker-compose --version
```

### First-time setup

```bash
git clone https://github.com/nima-abdpoor/brain-blitz.git
cd brain-blitz

# Create the shared Docker network (only once)
docker network create bb-network

# Start all infrastructure and services
docker-compose up -d
```

## Common Commands

| Task | Command |
|---|---|
| Start all services | `docker-compose up -d` |
| Stop all services | `docker-compose down` |
| View logs for a service | `docker-compose logs -f bb-game-service` |
| Run all tests | `go test ./...` |
| Run tests with verbose output | `go test -v ./...` |
| Format code | `gofmt -w .` |
| Run vet | `go vet ./...` |
| Install Docker | `make install-docker` |
| Install Docker Compose | `make install-docker-compose` |
| Install sql-migrate CLI | `make install-sql-migrate` |
| Build auth service binary | `go build -o auth-service ./cmd/auth/` |
| Build all binaries | `go build ./...` |

## Development Workflow

1. **Pick a service** — changes to a service live entirely in `services/<name>_app/`, `cmd/<name>/main.go`, and any shared `pkg/` or `adapter/` code it touches.
2. **Write failing test** — add to `service/<name>_test.go` or `delivery/http/<name>_test.go`.
3. **Implement** — in `service/service.go` (business logic) or `repository/<name>.go` (data access).
4. **Wire if needed** — add new dependencies to `app.go` constructor.
5. **Run tests** — `go test ./services/<name>_app/...`
6. **Open PR** → target `develop` → CI runs `go build ./...` and `go test -v ./...`.

## Debugging

### Log output

All services write structured JSON logs to stdout and to a rotating file under `logs/<service>/service.log` inside the container working directory.

To tail logs:
```bash
docker-compose logs -f bb-game-service
```

### Debug print in Auth Service

> **Warning**: `services/auth_app/service/service.go:62` contains `fmt.Println(svc.config.SecretKey, signedString, err)`. This prints the JWT secret key to stdout on every token creation. Remove before production deployment.

### Testing the WebSocket endpoint

Use `wscat` or any WebSocket client:
```bash
wscat -c "ws://localhost/game-service/api/v1/process-game" \
  -H "Authorization: Bearer <token>"
```

Send commands as JSON:
```json
{"command":"ADD_TO_WAITING_LIST","category":"SPORT"}
```

### Kafka

Kafka runs in KRaft mode (no ZooKeeper). Connect to the container:
```bash
docker exec -it kafka kafka-topics --bootstrap-server localhost:9092 --list
docker exec -it kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic matchMaking_v1_matchUsers --from-beginning
```

### MongoDB

Connect to the replica set primary:
```bash
docker exec -it bb-mongo1 mongosh --eval "rs.status()"
```

### Redis

```bash
docker exec -it bb-redis redis-cli
> KEYS *
> ZRANGE waiting_users:SPORT 0 -1 WITHSCORES
> GET game_questions_<gameId>
```

## Deployment

### Docker Compose (Development)

All services share the `bb-network` Docker network. Traefik listens on port 80 (HTTP).

Environment variable overrides for docker-compose are set inline in `docker-compose.yml` under each service's `environment:` block.

### Kubernetes

See `infra/kubernetes/` for all manifests. Deployment order:
1. `kubectl apply -f infra/kubernetes/volume/` — secrets, configmaps, PVCs
2. `kubectl apply -f infra/kubernetes/deployment/` — pods and services
3. `kubectl apply -f infra/kubernetes/ingress/` — Traefik IngressRoutes

The Auth Service secret key is stored in a Kubernetes Secret `auth-secret` and injected as `AUTH_service__secret_key`.

Docker images are pulled from `ghcr.io/nima-abdpoor/brain-blitz/<service>:latest`.

> Note: The Question Service does **not** have a Kubernetes deployment manifest yet.

## CI/CD

Three GitHub Actions workflows in `.github/workflows/`:

| Workflow | Trigger | Action |
|---|---|---|
| `go.yml` | Push/PR to `develop` | `go build ./...` + `go test -v ./...` |
| `docker-image-develop.yml` | Push/PR to `develop` (Dockerfile changed) | Build Docker images for auth, match, game, user (push=false) |
| `auto-tag.yml` | Push to `develop` | Auto-increment patch version tag and push |

## Important Files

| File | Purpose |
|---|---|
| `docker-compose.yml` | Full local environment definition |
| `go.mod` | Module name (`BrainBlitz.com/game`) and dependencies |
| `Makefile` | Helper targets for Docker and migration tooling |
| `prometheus.yml` | Prometheus scrape config (MongoDB exporter only) |
| `buf.yaml` / `buf.gen.yaml` | Protobuf toolchain config |
| `infra/deploy/<svc>/development/config.yaml` | Default config for each service |
| `contract/event/events.go` | Kafka topic name constants |
| `pkg/err_app/errors.go` | Error type definitions and sentinel errors |
| `pkg/cfg_loader/cfg_loader.go` | Config loading logic |

## Adding a New Service

1. Create `cmd/<name>/main.go` following the existing pattern.
2. Create `services/<name>_app/` with `app.go`, `config.go`, `delivery/`, `service/`, `repository/`.
3. Add a Dockerfile at `infra/deploy/<name>/development/Dockerfile`.
4. Add a config file at `infra/deploy/<name>/development/config.yaml`.
5. Add to `docker-compose.yml` with Traefik labels.
6. Add to the CI matrix in `docker-image-develop.yml`.
7. Add Kubernetes manifests in `infra/kubernetes/`.

## Adding a New Kafka Topic

1. Add the topic constant to `contract/event/events.go`.
2. In the producing service, call `broker.Publish(ctx, event.YOUR_TOPIC, payload)`.
3. In the consuming service, add the topic to the consumer's `getTopics()` method.
4. Define and add the Protobuf message if needed; regenerate Go code.
