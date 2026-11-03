# Deployment validation

Prepared on 2026-11-03 from repository commit `343a2b3` on local branch
`codex/deployment-readiness`. The source changes have not been pushed to GitHub.
Use the prepared source archive or transfer this working tree, rather than
cloning the original repository and expecting these changes to be present.

## Result

The Docker Compose backend has been built and exercised as a single-computer
deployment. Native runtime validation used ARM64 Linux containers under Docker
Desktop 29.9.2, with approximately 8 GB allocated to Docker. All five service
entry points also cross-compiled for Linux amd64. This is not a runtime test of
the user's other computer; its specifications have not yet been provided.

| Check | Result |
|---|---|
| Default Compose and smoke override validation | PASS |
| Native builds of all five application images | PASS |
| Full repository tests with the Go race detector | PASS |
| Final changed Game packages and smoke client, with race detection | PASS |
| `go vet` for all packages, then final changed packages | PASS |
| Linux amd64 and arm64 builds of all five service entry points | PASS |
| Fresh secret generation, permissions, and preserving existing `.env` | PASS |
| All five live service health checks through the gateway | PASS |
| Missing/invalid tokens and spoofed identity headers rejected | PASS |
| Internal token-minting route unavailable through the gateway | PASS |
| Two users: signup, login, profile, and authenticated WebSockets | PASS |
| Kafka matchmaking and MongoDB game creation | PASS |
| Ten seeded questions, concurrent readiness, stored answers, positive scores | PASS |
| Timed completion and finished players reconnecting | PASS |
| Stop/recreate deployment without deleting volumes | PASS |
| Existing accounts, 30 questions, and game history survive restart | PASS |
| PostgreSQL 17.11 starts against the initial PostgreSQL 17.2 data volume | PASS |
| Game smoke test after restart | PASS |
| API reachable from host at `http://127.0.0.1:8088` | PASS |

The first full-game test exposed an incorrect completion delay. It was fixed
and retested. Final unit regressions also cover players who submitted no
answers, disconnected players, old-socket cleanup, concurrent READY updates,
and READY arriving before roster initialization.

## Changes made

- Added the missing Question service, its database initializer, and repeatable
  demo question import.
- Standardized config paths in source and Docker images, corrected all Redis
  container addresses, and updated the previously empty Question Dockerfile.
- Added health checks, startup ordering, Kafka topic initialization, writable
  Kafka storage, repeatable MongoDB replica-set initialization, and named volumes.
- Replaced obsolete Bitnami runtime images with upstream Kafka and Redis images;
  selected the PostgreSQL 17 maintenance series.
- Added generated machine-local secrets, a private loopback gateway, internal
  infrastructure ports, unprivileged application containers, and build exclusions
  for secrets and logs. Monitoring is omitted from the default local stack.
- Corrected JWT expiry timestamps and added Bearer-token support and signing
  method checks. Token minting remains internal behind the Compose gateway.
- Corrected the game timer, persisted finished status, reset the full player
  roster, handled disconnected players, and made Redis readiness updates atomic.
- Corrected the HTTP response after WebSocket upgrade and protected replacement
  connections from stale cleanup.
- Added runnable checks, a two-player smoke client, a GitHub readiness workflow,
  and the guide [Run on another computer](run-on-another-computer.md).

Go dependency downloads required a reachable public mirror on this network.
The generated `.env` includes a fallback chain; Go checksum verification remains
enabled. No TLS verification bypasses were added.

## Remaining scope

- The target computer's OS, CPU capabilities, RAM allocation, and network access
  must be confirmed there. AMD64 binaries compile, but a full AMD64 runtime was
  not exercised on this ARM Mac.
- A Telegram bot implementation is not present in this repository. Bot startup,
  Telegram connectivity, identity mapping, and webhook/polling configuration
  require that client's source and settings. No Telegram messages were sent.
- The Question Add API remains a stub; the supplied SQL import provides demo
  questions. Import your real question data before using it beyond the demo.
- Kubernetes manifests, public HTTPS deployment, load/capacity limits, and outage
  recovery have not been validated. The Kafka consumer does not replay missed
  events or support horizontal service replicas; active games are not restored
  automatically across restarts.
- The added GitHub workflow is provided for repeatable checks but has not been
  executed on GitHub as part of this local preparation.

The source archive excludes `.env`, `.git`, IDE files, logs, and Docker data.
It includes `.env.example`. Generate new secrets on the destination computer.
