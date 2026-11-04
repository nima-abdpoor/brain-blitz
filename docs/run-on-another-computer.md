# Run BrainBlitz on another computer

This guide covers the Docker Compose deployment on one computer. Kubernetes
manifests are a separate deployment path and are not used by these instructions.

## Requirements

- A 64-bit Linux, macOS, or Windows computer. On Windows, use Docker Desktop with
  Linux containers and WSL 2; run the commands from your WSL terminal.
- Docker Engine/Desktop with Docker Compose v2.24 or newer.
- Python 3 for generating machine-local secrets. Go is built inside Docker.
- For a small demo, allow Docker **8 GB RAM**, at least **4 CPU cores**, and
  **20 GB free disk**. These are starting allocations, not measured capacity
  guarantees. A computer with 16 GB total RAM is preferable when also running
  other applications. MongoDB 8 requires a compatible modern CPU (including AVX
  on Intel/AMD); very old machines may not work.
- Internet access to pull images, download Go modules, and reach Telegram.

## Move the source

Copy the prepared source archive to the other computer and extract it. The
archive includes the readiness changes; cloning the original GitHub repository
alone will not include unpushed changes made during this preparation.

Do not copy this computer's `.env`, Docker socket, or logs. Generate a new `.env`
on the destination. If migrating existing game data, back up and restore the
databases separately and retain their original database credentials.

## First start

From the extracted project directory:

```sh
python3 scripts/init-env.py
docker compose config --quiet
docker compose --parallel 2 build
docker compose up -d --wait --wait-timeout 300
docker compose --profile tools run --rm seed-demo
docker compose ps -a
```

`init-env.py` does not overwrite existing secrets. The first build may take
several minutes. `seed-demo` is optional and repeatable: it inserts 10 EASY
questions for each of TECH, SPORT, and MUSIC without replacing existing rows.
Games need question data; the existing AddQuestion HTTP handler is unfinished,
so use the SQL import or your own database import until that feature is built.

The API gateway is available at **http://127.0.0.1:8088** by default. The `.env`
file controls `HTTP_BIND` and `HTTP_PORT`. Databases, Kafka, service ports, and
the Traefik admin dashboard are not published to the host network.

Check an endpoint:

```sh
curl --fail http://127.0.0.1:8088/user-service/public/api/v1/health-check
```

Other health endpoints:

- `/auth-service/api/v1/health-check`
- `/match-service/api/v1/health-check`
- `/game-service/api/v1/health-check`
- `/question-service/api/v1/question/health-check`

Users sign up and log in via `/user-service/public/api/v1/signup` and `/login`.
Use `Authorization: Bearer <accessToken>` for protected requests. Auth token
minting endpoints are intentionally internal: the User service calls Auth over
gRPC after checking credentials. They are not exposed by the gateway.

## Telegram bot

The `telegram-client` branch includes a bot in `clients/telegram-bot/`. Add
`BOT_TELEGRAM__BOT_TOKEN=<your BotFather token>` to the private `.env`, then run:

```sh
docker compose --profile telegram up -d --build bb-telegram-bot
docker compose --profile tools run --rm --no-deps seed-demo
docker compose logs --tail 30 bb-telegram-bot
```

Open the bot in Telegram and send `/start`. Register two players and select the
same category to play. The bot uses long polling and Redis session storage; it
does not require exposing the local backend to Telegram users. See
`clients/telegram-bot/README.md` for the full-game integration test.

If running the bot separately, use the appropriate backend base URL:

- Bot running directly on the same computer: `http://127.0.0.1:8088`.
- Bot running in a container on the same Compose network: `http://bb-traefik`.
- Bot running elsewhere: it needs a reachable address for this computer. A
  remote process cannot access this computer through its own `localhost`.

For a conventional chat bot, **long polling** is the easiest local setup: the
bot makes outbound connections to Telegram and outbound calls to the local
backend. Telegram users do not need direct access to your backend. Keep the
computer awake, online, Docker running, and the bot process running.

For **webhooks**, Telegram must reach the bot's HTTPS webhook endpoint. A tunnel
can expose that specific bot endpoint. Do not expose the Docker socket or
database ports. Telegram Mini Apps are different: browser clients need a public
HTTPS frontend and a reachable API. A local-only API will not work from users'
phones.

The bot also connects to the authenticated WebSocket at
`/game-service/api/v1/process-game`. Starting only the default Compose services
does not start the bot; enable the `telegram` profile as shown above.

## Run the checks

```sh
docker compose --profile tools build checks smoke
docker compose --profile tools run --rm checks
```

The check container runs race-enabled Go tests, `go vet`, and cross-compiles the
five service entry points for Linux amd64 and arm64.

The end-to-end smoke client creates two disposable accounts and plays one game.
Use an isolated deployment with short timers, not an existing game server:

```sh
docker compose -p brainblitz-validation -f docker-compose.yml -f compose.smoke.yaml up -d --build --wait --wait-timeout 300
docker compose -p brainblitz-validation --profile tools run --rm seed-demo
docker compose -p brainblitz-validation --profile tools run --rm smoke
docker compose -p brainblitz-validation down
```

Stop the ordinary deployment first or choose another `HTTP_PORT` to avoid a
host port conflict. The override is used for startup only; the seed and smoke
clients connect to the same named project. `down` retains volumes and data.

## Stop, restart, and preserve data

```sh
docker compose stop
docker compose start
docker compose logs --tail 100 bb-game-service
```

Postgres, Redis, Kafka, and each MongoDB node use named Docker volumes. Ordinary
`stop`, `start`, and `down` retain them. **`down -v` deletes the deployment's
database volumes.** Changing the generated Postgres password does not change a
password inside an already initialized database; retain `.env` across restarts.

Create SQL backups with `docker compose exec -T bb-postgres pg_dump -U user_admin
-d user_db` and repeat for `question_db`. Store the output securely. MongoDB
game history needs a `mongodump` backup too. Copying the source alone does not
copy database volumes.

## Network and startup troubleshooting

- **Image pulls fail:** verify internet/Docker registry access on that computer.
- **Go module proxy returns 403:** the generated `.env` tries the official Go
  proxy, then the public `goproxy.cn` mirror, then upstream source. You can edit
  `GOPROXY` to select a reachable proxy or use `direct`. Module checksums remain
  enabled; do not disable TLS verification or checksum validation. Building
  directly without the generated `.env` uses the Dockerfile's official-proxy
  and upstream-source fallback instead.
- **Container is unhealthy:** inspect `docker compose ps -a` and that service's
  logs. Initializer jobs should exit with code 0; they do not run continuously.
- **No matchmaking:** use two distinct accounts in the same category. The
  scheduler runs every 15 seconds by default.
- **No questions:** run the demo import or load EASY questions for the selected
  category in `question_db`.
- **Bot cannot reach Telegram:** check `https://api.telegram.org` from the target
  computer and the bot process/container. This computer's VPN configuration does
  not transfer with the project.

## Scope and operational limits

This setup is intended for a trusted single-computer demo and development. Keep
the default loopback binding for a bot on the same machine. Before publishing
the API to the internet, configure TLS, access controls, rate limiting, and
backups appropriate to the deployment. Redis, MongoDB, and Kafka are reachable
only on the private Compose network and do not have application authentication;
only trusted containers should join that network. The gateway's Docker provider
uses a Docker socket mount; do not give untrusted applications access to it.

The Kafka adapter currently consumes only partition 0 at the newest offset and
does not use consumer groups. Run one instance per application service. It
does not replay missed game events after an outage. Active WebSocket sessions and
in-progress games are not restored automatically after a backend restart. Three MongoDB containers on
one computer provide a replica set for application operations, not resilience
against that computer failing. The bot client and the actual destination
hardware must still be checked on the destination before calling the deployment
fully verified there.
