# BrainBlitz Telegram Bot

A Telegram bot client for [BrainBlitz](../../README.md), the real-time multiplayer quiz game backend. Players register, log in, join a category queue, and play a live timed quiz — entirely through Telegram chat — while this process holds the WebSocket connection to the game backend on their behalf.

This is its own Go module (`github.com/nima-abdpoor/brain-blitz-telegram-bot`), independently versioned and deployed from the main BrainBlitz backend, though it lives inside the same monorepo (`clients/telegram-bot/`).

## Features

- **Onboarding**: `/start`, inline register/login conversations
- **Profile & menu**: `/profile`, `/help`, logout
- **Matchmaking**: join a category queue (`SPORT` / `MUSIC` / `TECH`), cancel, match-found notification
- **Gameplay**: server-controlled rounds advance when everyone answers or time expires; full-width answer buttons, local countdown edits, named standings, and personalized winner/loser/draw results
- **Resilience**: retrying HTTP client with backoff, WebSocket-connect retry, optional rate limiting, optional Redis-backed session persistence across restarts
- **Observability**: Prometheus metrics, structured JSON logging (see [Metrics](#metrics) below)

See `docs/client/` in the repo root for the full design history (architecture, business flows, phased roadmap) behind this implementation.

## Architecture

```
internal/core/{auth,session,profile,matchmaking,game}   Telegram-agnostic business logic
        │
internal/apiclient/{http,ws}                            Typed BrainBlitz API clients
        │
internal/telegram/{commands,conversation,keyboards,middleware}   Delivery layer (owns Telegram SDK types)
```

`internal/core` and `internal/apiclient` know nothing about Telegram — a future Web/Discord/CLI client would reuse them and add its own delivery adapter. `internal/telegram` is the only layer allowed to import the Telegram Bot API SDK.

A single bot process serves every Telegram chat; it multiplexes many per-chat WebSocket connections to game-service, guarded by a per-chat state machine (`internal/core/matchmaking`, `internal/core/game`) since there is no server-side session/reconnect API.

## Requirements

- Go 1.23 (the module pins `go 1.23.0` deliberately — see `go.mod`)
- A running BrainBlitz backend (`docker-compose up` from the repo root) reachable via its Traefik gateway
- A Telegram bot token from [@BotFather](https://t.me/BotFather)
- Redis, only if you opt into `session_store.type: redis` (see [Configuration](#configuration))

## Getting Started

From the repository root, add `BOT_TELEGRAM__BOT_TOKEN=<your BotFather token>`
to the private, ignored `.env`, then run:

```sh
docker compose --profile telegram up -d --build bb-telegram-bot
docker compose --profile tools run --rm --no-deps seed-demo
docker compose logs --tail 30 bb-telegram-bot
```

Open the bot in Telegram and send `/start`. Players choose a category, timer,
and game size (2–15 by default). A match starts when the selected number of
registered players choose the same settings. The Compose bot uses the internal gateway and persists authentication sessions in Redis. Keep one bot
instance running per token. Stop it with `docker compose stop bb-telegram-bot`.

If Telegram is unreachable from Docker, connect the computer through a network
that can reach `api.telegram.org`, or set `BOT_HTTPS_PROXY` in `.env` to an
existing proxy reachable from the container. Use `host.docker.internal` instead
of `127.0.0.1` for a proxy running on a Docker Desktop host. Backend traffic
stays on the internal Compose network.

To run directly with a local Go installation instead:

```sh
cd clients/telegram-bot
go build ./...

export BOT_TELEGRAM__BOT_TOKEN="<token from @BotFather>"
go run ./cmd/bot
```

By default the bot loads `configs/development/config.yaml`, pointed at `http://127.0.0.1:8088/user-service` etc. Override the config file path with `BOT_CONFIG_FILE`, or any individual key with a `BOT_`-prefixed environment variable (`__` separates nesting, e.g. `BOT_BACKEND__HTTP_TIMEOUT`).

## Configuration

All settings live in `configs/development/config.yaml`; every key can be overridden by an environment variable. Sections:

| Section | Purpose | Default |
|---|---|---|
| `telegram.bot_token` | Bot API token | required, no default |
| `backend.*` | Gateway URLs + HTTP timeout | `http://127.0.0.1:8088/...` |
| `session_store` | `memory` (lost on restart) or `redis` (persists auth sessions only) | `memory` |
| `rate_limit` | Caps outbound HTTP/WS-connect rate | unlimited |
| `metrics` | Prometheus `/metrics` endpoint | disabled |
| `logging.level` | `debug`/`info`/`warn`/`error` | `info` |

See the comments in `configs/development/config.yaml` for the full, current set of keys and their defaults.

## Metrics

Set `metrics.enabled: true` (or `BOT_METRICS__ENABLED=true`) to serve Prometheus metrics on `metrics.listen_addr` (default `:9090`) at `/metrics`:

- `bot_commands_total{command}` — Telegram commands handled
- `bot_callbacks_total{callback}` — inline-keyboard callbacks handled (per-question answer buttons are recorded by prefix, not full callback data, to avoid an unbounded label cardinality)
- `bot_http_request_duration_seconds{method,path}` — outbound HTTP latency per attempt, including retries
- `bot_http_request_errors_total{method,path}` — failed outbound HTTP attempts
- `bot_active_ws_connections` — currently open game-service WebSocket connections

This closes a gap in the main BrainBlitz backend itself, which wires Prometheus/Grafana in `docker-compose.yml` but has no service that actually exposes application metrics (`docs/context/07-known-issues.md` #7) — this bot does not repeat that omission.

## Structured Logging

The bot logs structured JSON to stdout (`internal/logger`) at a configurable level. Logging has been audited end to end: no call site logs a password, access token, or refresh token — auth flows only ever pass credentials through local variables for the duration of a single request. Container runtimes are expected to capture stdout; this process does not manage its own log files or rotation.

## Testing

```sh
go test ./...                       # unit tests (fakes only, no network) — always safe to run
go test ./... -race                 # same, with the race detector

docker-compose up -d                # from the repo root — brings up the real backend
go test -tags=integration ./integration/...   # exercises real signup/login/profile and a live 2-player match
```

Integration tests each perform their own reachability check and call `t.Skip` (not fail) if the backend they need isn't up, so `-tags=integration` is safe to run even without `docker-compose up`.

The full-game test drives 2, 3, 4, and 15 simulated Telegram chats through the real
dispatcher, registration/login conversations, answer keyboards, backend
services, and final leaderboard. Only Telegram's transport is replaced; this
test does not verify delivery through Telegram's servers or send real messages.
Run it against an isolated deployment with short timers, from the repo root:

```sh
HTTP_PORT=8089 docker compose -p brainblitz-telegram-e2e -f docker-compose.yml -f compose.smoke.yaml up -d --build --wait --wait-timeout 300
docker compose -p brainblitz-telegram-e2e --profile tools run --rm --no-deps seed-demo
docker compose -p brainblitz-telegram-e2e --profile tools run --rm --no-deps -e BOT_INTEGRATION_FULL_GAME=1 telegram-tests
docker compose -p brainblitz-telegram-e2e --profile tools run --rm --no-deps -e BOT_INTEGRATION_FULL_GAME=1 -e BOT_INTEGRATION_SILENT_PLAYER=1 telegram-tests go test -race -v -tags=integration -count=1 -timeout=12m -run TestIntegration_TelegramConversationAndFullGame ./integration/...
docker compose -p brainblitz-telegram-e2e down
```

## Deployment

See [deploy/README.md](deploy/README.md) for the Docker image and Kubernetes manifests, and the constraints specific to a long-polling Telegram bot (in particular: **do not scale replicas past 1** without first switching to webhook mode).

## Known Limitations

See `docs/client/client-roadmap.md` and `docs/client/gap-analysis.md` in the repo root for the full list; the most consequential for day-to-day operation:

- No backend token-refresh endpoint exists — a session simply expires and the user is asked to `/login` again (`session.Session.Stale`, 24h).
- A dropped WebSocket connection (queued or mid-game) is never auto-resumed; the bot has no way to know server-side truth after reconnecting, so it always ends the attempt locally rather than risk double-enqueuing or missing an already-fired match.
- Categories are hardcoded client-side (`SPORT`/`MUSIC`/`TECH`) since the backend's `GET_CATEGORIES` command is a documented no-op.

## Round protocol and operational behavior

The backend publishes one active question at a time, with `roundNumber`,
`totalQuestions`, and an absolute deadline in `QUESTIONS_PUBLISHED`. Each
accepted answer counts once, including incorrect answers. The server advances
all players together after all answers arrive or the deadline expires, and
finishes after the final round. The bot continues to understand the older
all-questions-at-once response for compatibility with older backends.

The bot uses a player's Telegram first/last name in standings; account emails
are not used as a name fallback. The countdown runs locally and edits only the
Telegram message every five seconds. It does not poll the backend. A slow
chat's requests do not block other chats. Telegram requests have separate
bounds for callbacks (2 seconds), ordinary requests (8 seconds), and polling
(40 seconds). Explicit transient API rejections may be retried within a bounded
budget; ambiguous network errors during message delivery are logged rather
than automatically duplicated. External network outages can still interrupt
Telegram delivery.

Queue cancellation and socket closure clear shared backend presence and queue
membership. A player who disconnects after a match has formed is treated as
unanswered for its remaining readiness/round deadlines. Authentication survives
bot restarts in Redis; in-flight game recovery remains unsupported. These
services still run as a single instance per backend service.


### Profile game statistics

`/profile` and the Profile button include completed games, wins/losses/draws,
win rate, lifetime points, average/best score, submitted/correct answers,
accuracy, and global rank. Rank uses lifetime points from completed games;
ties share a competition rank (1, 1, 3). Players with no completed games are
unranked. Historical finished games are included automatically; unfinished
games are excluded. A statistics outage leaves account details available and
shows a temporary-unavailability message rather than zero totals.

The user service fetches these fields from the game service's
`GET /api/v1/statistics` using the authenticated user ID. Through the gateway,
this endpoint is `/game-service/api/v1/statistics` and requires authentication.
Set `USER_game_service_url` to the internal game-service HTTP base URL;
Compose and Kubernetes manifests provide it. Development defaults to
`http://127.0.0.1:5003`. Service ports must remain internal, as with existing
routes that consume the gateway's `X-User-ID` header.

### Question timers and round recaps

New-game flow: choose a category, question time limit, and total player count,
then wait for players with the same category, timer, and game size. Player counts come from
`GET /game-service/api/v1/game-options`; configure the backend with `MIN_PLAYERS`
and `MAX_PLAYERS` in Compose (defaults 2 and 15), and set
`BOT_BACKEND__GAME_SERVICE_URL` to the HTTP game-service base URL when deploying
the bot separately. The bot loads timer
choices from the backend; PostgreSQL initially seeds **10 seconds** and
**1 minute**. Configure `BOT_BACKEND__QUESTION_SERVICE_URL` for the gateway's
question-service base URL (provided in Compose/development configuration).
The game service uses `GAME_question_service_url` for its internal validation.

Once everyone answers or the question expires, each player receives their own
answer and the correct answer before the next question. A missed answer is
shown explicitly. The final question also gets a recap before game results.
Solutions are not included in early answer acknowledgements.

Admin/backoffice API (all paths are relative to `/question-service`):

- `GET /api/v1/question/time-options`: enabled options, authenticated players.
- `GET /api/v1/question/admin/time-options`: all options, admin role required.
- `PUT /api/v1/question/admin/time-options/{id}`: create/update an option,
  admin role required. Body example:
  `{"label":"10 seconds","seconds":10,"enabled":true,"sortOrder":10}`.

IDs use 1–32 letters, numbers, underscores, or hyphens; durations allow 5–3600
seconds. Disable an option with `enabled:false`. Edits apply to future queue
joins; an existing queued selection or active game keeps its duration snapshot.
The migration's seed uses `ON CONFLICT DO NOTHING`, preserving later edits.
No standalone backoffice UI exists in this repository. Management APIs require
an admin token at the gateway; service ports remain internal.

Older clients that omit `timeOptionId` retain the configured legacy timeout
and match only other legacy clients. The updated bot has no preselected timer
and requires an explicit choice. A configuration outage or empty option list
shows a retry message instead of silently choosing a duration.
