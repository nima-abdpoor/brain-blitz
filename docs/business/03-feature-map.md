# Feature Map

## Authentication & Identity

| Feature | Status | Notes |
|---|---|---|
| User registration (email + password) | Complete | `POST /user-service/public/api/v1/signup` |
| User login (returns access + refresh tokens) | Partial | Login works; refresh token is actually an access token (bug R-01) |
| JWT access token (HS256, 24h) | Complete | `POST /auth-service/api/v1/access-token` |
| JWT refresh token (HS256, 120h) | Partial | Bug: returns access token instead of refresh token |
| Token validation (Traefik ForwardAuth) | Complete | `GET /auth-service/api/v1/validate-token` |
| User profile retrieval | Complete | `GET /user-service/api/v1/profile` |
| Token refresh flow (use refresh token to get new access token) | Missing | No endpoint to exchange refresh → access token |
| User logout / token revocation | Missing | JWT is stateless; no blacklist |
| Admin role management | Partial | Role field exists; no admin-only endpoints |
| Password change | Missing | No endpoint |
| Email verification | Missing | No email sending infrastructure |

## Matchmaking

| Feature | Status | Notes |
|---|---|---|
| Join matchmaking queue via WebSocket | Complete | `ADD_TO_WAITING_LIST` command |
| Join matchmaking queue via HTTP | Complete | `POST /match-service/api/v1/addToWaitingList` |
| Per-category waiting lists | Complete | Redis sorted sets per category |
| FIFO matching (oldest waiters first) | Complete | Sorted by join timestamp |
| Periodic matchmaking scheduler | Complete | Every 15 seconds |
| Even-number pairing (pairs, quads) | Complete | Drops most recent if odd |
| 20-minute waiting list TTL | Complete | `min_time_list_selection: -20m` |
| Match notification to players | Complete | `MATCH_CREATED` WebSocket event |
| Multiple players per match (2+) | Partial | Logic supports N but UI/game assumes 2 |
| Player kicked from queue after timeout | Missing | No eviction of stale entries (only window filter) |
| Queue depth visibility | Missing | No endpoint to query queue size |

## Game Session

| Feature | Status | Notes |
|---|---|---|
| WebSocket game session | Complete | `GET /game-service/api/v1/process-game` |
| Game document creation in MongoDB | Complete | On `MATCH_CREATED` consume |
| Question delivery to players | Complete | `QUESTIONS_PUBLISHED` event with all questions upfront |
| Per-question answer deadline | Complete | `ValidAnswerTime` set on READY |
| Player ready synchronization | Partial | Bug: game starts 1 player early (R-04) |
| Answer submission | Complete | `ANSWER` command |
| Duplicate answer prevention | Complete | Checked against MongoDB |
| Anti-cheat (too-early answer rejection) | Complete | Answer rejected before valid window |
| Scoring per answer | Complete | BaseScore + time bonus |
| Real-time leaderboard after each answer | Complete | Returned in ANSWER_ACCEPTED event |
| Game auto-completion via timer | Complete | Asynq deferred task |
| Final leaderboard delivery | Complete | `COMPLETED` WebSocket event |
| Game status update on completion | Missing | MongoDB game document status not updated to FINISHED |
| Reconnection handling | Missing | No WebSocket reconnect flow |
| Spectator mode | Missing | |
| Game replay / history | Missing | No retrieval API for past games |
| Multiple rounds per game | Missing | Single question set, no round concept |

## Questions

| Feature | Status | Notes |
|---|---|---|
| Question storage (PostgreSQL) | Complete | `questions` table with UUID PK |
| Random question selection by category | Complete | `GetRandomQuestions` |
| Question difficulty filtering | Partial | `GetRandomQuestions` takes difficulty param but `ConsumeMatchCreated` always passes `EASY` |
| History-aware question selection | Partial | `GetProperQuestions` implemented but never called; `user_question_history` never populated |
| Add question via API | Missing | `AddQuestion` is a stub |
| Question search / listing | Missing | No admin API |
| Question categories: SPORT, MUSIC, TECH | Complete | Hardcoded enum |
| 4-choice questions | Complete | `choices TEXT[]` |

## Infrastructure

| Feature | Status | Notes |
|---|---|---|
| Docker Compose local environment | Complete | All services + infra |
| Kubernetes manifests | Partial | Missing Question Service deployment |
| MongoDB replica set (3 nodes) | Complete | `rs0` |
| Kafka event streaming | Complete | Match, Game, Question |
| Redis for game state and queue | Complete | |
| Prometheus scraping | Partial | MongoDB only; no app metrics |
| Grafana dashboards | Missing | No pre-configured dashboards |
| Structured JSON logging | Complete | slog + lumberjack |
| Graceful shutdown | Complete | All services |
| CI: build + test | Complete | GitHub Actions on develop |
| CI: Docker image build | Partial | Build only, push=false |
| Auto version tagging | Complete | Patch increment on develop push |
| Application-level metrics | Missing | 15+ todo comments |
| Distributed tracing | Missing | No OpenTelemetry |
| Health check endpoints | Complete | All services |

## Security

| Feature | Status | Notes |
|---|---|---|
| Password hashing (bcrypt) | Complete | |
| JWT authentication | Complete | |
| Route-level auth via Traefik ForwardAuth | Complete | |
| CORS configuration | Partial | Configured as `*` (allow all) |
| Secret management (K8s Secrets) | Partial | Auth secret only; others use configmaps |
| Rate limiting | Missing | |
| Input sanitization beyond validation | Unknown | |
| HTTPS/TLS | Missing | No TLS termination configured |
