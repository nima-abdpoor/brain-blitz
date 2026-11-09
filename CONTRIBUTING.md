# Contributing to BrainBlitz

Contributions to the backend, Telegram client, tests, and documentation are
welcome. For suspected security vulnerabilities, follow [SECURITY.md](SECURITY.md)
instead of including details in a public issue or pull request.

## Before starting

Search existing [issues](https://github.com/nima-abdpoor/brain-blitz/issues) and
[pull requests](https://github.com/nima-abdpoor/brain-blitz/pulls) for related work.
For a large feature, protocol change, or architectural change, open an issue
describing the problem and proposed approach before investing in implementation.
Small fixes and documentation improvements can go straight to a pull request.

Keep discussions respectful, describe observable behavior, and explain technical
disagreements with examples. Review comments should help improve the change.

## Fork and create a branch

Fork the repository on GitHub, then create a branch from the upstream default
branch, currently `main`:

```sh
git clone https://github.com/YOUR-USERNAME/brain-blitz.git
cd brain-blitz
git remote add upstream https://github.com/nima-abdpoor/brain-blitz.git
git fetch upstream
git switch -c feat/short-description upstream/main
```

Use a descriptive branch name such as `fix/queue-cancellation` or
`docs/local-setup`. Target `main` when opening the pull request unless a
maintainer requests another base branch.

## Set up a local environment

The standard local stack uses Docker Compose v2.24 or newer and Python 3.
The Docker build supplies Go; a native Go installation is optional. If you
develop without the build container, check both modules' `go.mod` files and use
a compatible Go toolchain. The Compose build currently uses Go 1.26.

Follow [Redis host setup](docs/redis-host-setup.md), then run from the repository
root:

```sh
python3 scripts/init-env.py
docker compose config --quiet
docker compose --parallel 2 build
docker compose up -d --wait --wait-timeout 300
docker compose --profile tools run --rm seed-demo
curl --fail http://127.0.0.1:8088/user-service/public/api/v1/health-check
```

The generated `.env` is private and ignored by Git. Do not commit credentials or
include them in issue reports. Document new configuration keys in `.env.example`
or the relevant service configuration using empty or non-secret examples.

The Telegram bot is optional. To test actual Telegram delivery, configure your
own test bot token in `.env` and follow the [bot guide](clients/telegram-bot/README.md).
Unit tests and the simulated Telegram integration tests do not need a Telegram
token. Use one bot process per token.

See the [deployment guide](docs/run-on-another-computer.md) for resource
requirements, connectivity, and troubleshooting. `docker compose stop` stops
the local stack while retaining its data.

## Work in the right module and layer

This repository contains two independent Go modules:

| Location | Responsibility |
| --- | --- |
| Repository root | Backend services, infrastructure adapters, shared packages, and smoke client |
| `clients/telegram-bot/` | Telegram client, backend API clients, and bot integration tests |

Running `go test ./...` from the root does **not** test the nested Telegram module.

- Keep transport handling in `delivery/`, business rules in `service/`, and
  persistence in `repository/`. Put infrastructure integrations in `adapter/`.
- Keep the bot's `internal/core/` independent of the Telegram SDK. Telegram
  message formatting and callbacks belong in `internal/telegram/`.
- Use `gofmt` on changed Go files. Follow existing package conventions, pass
  dependencies explicitly, and return useful errors.
- Preserve context cancellation, bounded network operations, and safe ownership
  of shared state and goroutines. Cover concurrent behavior with race tests.
- For database changes, add migrations in the owning service and describe
  compatibility or rollout requirements in the pull request.
- When adding Kafka events, update `contract/event/events.go`. When changing
  Protobuf contracts, preserve existing field numbers and commit regenerated
  Go files together with their `.proto` sources.
- Update relevant documentation and configuration examples when behavior changes.

Keep changes focused on the problem being addressed. Avoid mixing unrelated
formatting, dependency updates, or generated files into a feature or fix.

## Test your change

Start with tests for the affected packages. Add regression coverage for bug
fixes and cover new business rules, boundaries, and failure paths. Documentation
changes generally need link and formatting checks rather than application tests.

For backend changes, run from the repository root:

```sh
go test -race ./...
go vet ./...
```

For Telegram client changes, also run in its module:

```sh
cd clients/telegram-bot
go test -race ./...
go vet ./...
```

To run the backend checks with the build toolchain instead of a native Go
installation, use the following from the repository root:

```sh
docker compose --profile tools build checks
docker compose --profile tools run --rm checks
```

This runs the checks in [scripts/check.sh](scripts/check.sh): formatting checks,
race tests, `go vet`, and Linux amd64/arm64 builds.

### Integration and multiplayer changes

Use a separate Compose project and port for integration tests. The smoke
override shortens timers; keep it separate from your normal local game server.
Tests create disposable accounts and game records.

Run the bot's full-game scenarios from the repository root:

```sh
HTTP_PORT=8089 docker compose -p brainblitz-contrib -f docker-compose.yml -f compose.smoke.yaml up -d --build --wait --wait-timeout 300
docker compose -p brainblitz-contrib --profile tools run --rm --no-deps seed-demo
docker compose -p brainblitz-contrib --profile tools run --rm --no-deps -e BOT_INTEGRATION_FULL_GAME=1 telegram-tests
docker compose -p brainblitz-contrib --profile tools run --rm --no-deps -e BOT_INTEGRATION_FULL_GAME=1 -e BOT_INTEGRATION_SILENT_PLAYER=1 telegram-tests go test -race -v -tags=integration -count=1 -timeout=6m -run 'TestIntegration_TelegramConversationAndFullGame/15_players$' ./integration/...
docker compose -p brainblitz-contrib down
```

These scenarios exercise real backend services with simulated Telegram
transport. They cover 2-, 3-, 4-, and 15-player games and a 15-player game with a
silent player. Tests that skip because the backend is unavailable do not count
as integration validation; inspect the output. The cleanup command retains
the test project's database volumes.

For matchmaking and gameplay changes, cover exact roster size, isolation by
category/timer/player count, incomplete queues, cancellation, concurrent
readiness and answers, timeouts, recaps, and final results as relevant.

The [deployment readiness workflow](.github/workflows/deployment-readiness.yml)
defines the automated backend, build, and multiplayer integration checks.

## Open a pull request

Use the pull request template. Explain the problem and resulting behavior,
link related issues, and include the commands you ran and their results.
Mention skipped checks, migrations, configuration changes, and compatibility
or rollout considerations when they apply. Include screenshots for changes to
user-visible flows when they help a reviewer understand the result.

Use a clear, descriptive commit message. Keep the pull request current with its
base branch, respond to review feedback, and ensure applicable CI checks pass.
If you need early feedback, open a draft pull request.

Contributions are made under the repository's [MIT license](LICENSE). Preserve
third-party notices and identify any new dependencies or borrowed code.
