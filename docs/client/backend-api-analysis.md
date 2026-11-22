# Backend API Analysis (Client Integration Perspective)

This reference describes the gateway and game protocol used by the Android app and Telegram bot. [Android integration](android-roadmap.md#existing-backend-contracts) · [Telegram setup](../../clients/telegram-bot/README.md)

Assumptions are marked **[ASSUMPTION]**.

---

## 0. Topology Recap

| Service | HTTP | gRPC | Client-reachable via gateway? |
|---|---|---|---|
| `auth` | `:5000` | `:6000` | Compose exposes the health check. Token operations are internal to ForwardAuth and the User service. |
| `user` | `:5001` | `:6001` (unused stub) | Yes — `/user-service/...` |
| `match` | `:5002` | none (skeleton only, never wired) | Yes — `/match-service/...` |
| `game` | `:5003` (HTTP + WS on same port) | n/a | Yes — `/game-service/...` |
| `question` | `:5004` | none | Compose routes `/question-service/api/v1/question/`. The checked-in Kubernetes ingress needs this route for timer selection. |

Android reads question timers through the Compose gateway. Confirm the same endpoint and authentication middleware on staging before building for a physical device. [Compose routes](../../docker-compose.yml) · [Kubernetes ingress](../../infra/kubernetes/ingress/ingress-route.yaml)

---

## 1. AUTH SERVICE — internal only

Clients sign in through User Service. The User service mints tokens over gRPC, and the gateway validates access tokens through ForwardAuth.

### `POST /api/v1/access-token`
- **Auth**: none (this endpoint mints tokens; it doesn't consume them)
- **Request**:
  ```json
  { "data": [ {"key": "id", "value": "123"}, {"key": "role", "value": "user"} ] }
  ```
- **Response 200**:
  ```json
  { "access_token": "string", "expire_time": 86400000 }
  ```
- **Errors**: `400` invalid body / missing `data`; `500` signing failure.
- **Purpose**: Mint a JWT access token from arbitrary claims.
- **Client feature**: None directly — only `user` service calls this today.
- **Expiry** uses an absolute Unix timestamp. Android reads the access token deadline and asks the player to sign in again when it expires. [Auth token implementation](../../services/auth_app/service/service.go)

### `POST /api/v1/refresh-token`
Same shape as above, mints a longer-lived token (`120h` in the development config).

### `POST /api/v1/validate-token` and `GET /api/v1/validate-token`
- **Auth**: token passed via `Authorization` header, not body.
- **Response 200**: `{ "valid": true, "data": [{"key":"id","value":"123"}, {"key":"role","value":"user"}] }`
- **Response 401**: `{ "valid": false }` (missing header, bad body, or invalid/expired JWT)
- **Purpose**: This is what Traefik's `forwardAuth` middleware calls on every protected request; it injects `X-User-ID`, `X-User-Role`, `X-Auth-Data` headers into the downstream request.
- **Client feature**. ForwardAuth validates authenticated Android and Telegram HTTP requests and WebSocket upgrades.

**gRPC** (`auth.TokenService`, port `6000`): `GetAccessToken`, `GetRefreshToken` — same payloads as above, called only by `user` service today.

---

## 2. USER SERVICE — `/user-service/...`

### `GET /user-service/public/api/v1/health-check`
Public. Returns `{ "message": "everything is good!" }`. No dependency checks performed (doesn't ping Postgres).

### `POST /user-service/public/api/v1/signup`
- **Auth**: none (public group)
- **Request**:
  ```json
  { "email": "alice@example.com", "password": "hunter2" }
  ```
- **Validation**: email must contain exactly one `@`; local-part length `> 4`; domain length `> 2` and `< 50` (`pkg/email/emailValidation.go` — a manual check, not RFC-compliant regex). Password just needs `len > 0` — **no minimum length, complexity, or format requirement**.
- **Response 200**: `{ "displayName": "alice" }` (derived from the email local-part; the user can't choose their own display name)
- **Errors**:
  | Status | Body | Cause |
  |---|---|---|
  | 400 | `{"message":"invalid username or password"}` | malformed JSON |
  | 400 | `{"message":"invalid input","error":"invalid input"}` | invalid email format |
  | 403 | `{"message":"invalid username or password"}` | empty password (note: mapped to the *login* error message even though this is signup) |
  | 400 | `{"message":"username already exists"}` | duplicate email |
  | 500 | `{"message":"something went wrong"}` | bcrypt or DB failure |
- **Purpose**: Register a new player. Role is always `"user"` — no way to create an admin via this API.
- **Client feature**: `/register` or first-run onboarding flow.

### `POST /user-service/public/api/v1/login`
- **Auth**: none
- **Request**: `{ "email": "...", "password": "..." }` (both required)
- **Response 200**: `{ "id": "123", "accessToken": "...", "refreshToken": "..." }`
- **Errors**:
  | Status | Body | Cause |
  |---|---|---|
  | 400 | `{"message":"invalid username or password"}` | malformed JSON |
  | 403 | `{"message":"invalid username or password"}` | invalid email format, empty password, OR wrong password |
  | 500 | `{"message":"something went wrong"}` | **user not found** (bug: a nonexistent email returns 500, not 401/403/404 — the repository error for "no such user" is wrapped as `ErrInternal`) |
  | 500 | `{"message":"something went wrong"}` | Auth Service gRPC unreachable |
- **Client-critical defect**: `refreshToken` returned here is a bug in `adapter/auth/client.go` — confirmed **fixed** per repo history (R-01), so as of the current `develop` branch this should be a real refresh token. A client should still verify by decoding the JWT and checking its expiry claim differs from the access token's, given the separate `exp`-encoding defect noted above.
- **Purpose**: Authenticate and receive tokens.
- **Client feature**: `/login`, session bootstrap.

### `GET /user-service/api/v1/profile` (protected — ForwardAuth)
- **Auth**: `Authorization: Bearer <accessToken>`; Traefik injects `X-User-ID` before this handler runs.
- **Request**: no body.
- **Response 200**:
  ```json
  { "id": "123", "username": "alice@example.com", "displayName": "alice", "role": "user", "createdAt": 1719999999000, "updatedAt": 1719999999000 }
  ```
- **Errors**:
  | Status | Body | Cause |
  |---|---|---|
  | 400 | `"Invalid user id"` (bare JSON string, not an object — inconsistent with every other error shape in the system) | missing `X-User-ID` header (shouldn't happen through the gateway, but possible if called directly) |
  | 404 | `{"message":"user not found"}` | user deleted after token issued |
  | 500 | `{"message":"something went wrong"}` | DB failure |
- **Purpose**: Fetch the caller's own profile.
- **Client feature**: `/profile` command, main menu display.

**gRPC** (port `6001`): registered but **no service is actually bound to the server** — `GetCustomer` exists as a bare Go method, not wired to any `.proto` contract. Treat as non-functional; do not build against it.

---

## 3. MATCH SERVICE — `/match-service/api/v1/...`

Entire router (including health-check) sits behind Traefik's `auth-middleware` in the current deployment — there is no public sub-router for match, unlike user-service.

### `GET /match-service/api/v1/health-check`
Protected (per current deployment config, unlike its name might suggest). Returns `{ "message": "everything is good!" }`.

### `POST /match-service/api/v1/addToWaitingList`
- **Auth**: `Authorization: Bearer <token>` (ForwardAuth) + handler also independently requires `X-User-ID`.
- **Request**: `{ "category": "SPORT" }` — must be exactly `SPORT`, `MUSIC`, or `TECH` (case-sensitive).
- **Response 200**: `{ "timeout": 1200000000000 }` — this is a raw Go `time.Duration` serialized as **nanoseconds**, not a human string or seconds. A client must divide by `1e9` to get seconds (20 minutes in dev config).
- **Errors**:
  | Status | Body | Cause |
  |---|---|---|
  | 400 | `"Invalid user id"` (bare string) | missing `X-User-ID` |
  | 400 | `"invalid category"` (bare string) | malformed JSON body |
  | 400 | `{"message":"invalid input"}` | category not in `{SPORT,MUSIC,TECH}` |
  | 500 | `{"message":"something went wrong"}` | Redis failure |
- **Purpose**: HTTP alternative to the WebSocket `ADD_TO_WAITING_LIST` command. **Important**: joining via this endpoint alone does not let the player receive `MATCH_CREATED` — that only arrives over the game-service WebSocket. A client must have a WS connection open regardless of which path it used to join the queue.
- **Client feature**: Not needed if the bot always uses the WebSocket path (recommended — see `client-architecture.md`).

No gRPC server is actually running for match-service despite config/proto scaffolding existing.

---

## 4. GAME SERVICE — `/game-service/api/v1/...` (HTTP + WebSocket, same port)

### `GET /game-service/api/v1/health-check`
Public in effect (reads but doesn't validate auth headers). `{ "message": "everything is good!" }`.

### `GET /game-service/api/v1/game-options`

An authenticated request returns `minPlayers`, `maxPlayers`, and `numberOfPlayers`. The default range is 2 to 15 inclusive. Android uses `numberOfPlayers` for its player-count selector. [Game options](../../services/game_app/service/options.go)

### `GET /game-service/api/v1/process-game` — **WebSocket upgrade endpoint**

This is the single most important endpoint for the client and the only way to actually play a game.

- **Auth**: `Authorization: Bearer <token>` for ForwardAuth (gateway-level) — **but the game-service handler itself only checks for a raw `X-User-ID` header**, with no token/signature validation of its own. In production this header only exists because Traefik injected it after ForwardAuth succeeded, but if a client could reach the game-service directly (bypassing the gateway) it could impersonate any user ID. This is an internal trust boundary a client developer should be aware of but cannot fix — connect through the gateway URL.
- **Handshake**: standard WebSocket upgrade (`ws://` or `wss://` scheme), no subprotocol negotiation used by the app.
- **Immediate server push after connect** for `UNKNOWN` or `FINISHED` user status. The server sends a raw JSON object with categories and allowed player counts. For example, a two-player-only deployment sends:
  ```json
  { "categories": ["SPORT", "MUSIC", "TECH"], "numberOfPlayers": [2] }
  ```
  Game recovery is opt-in. Send `SYNC_MATCH` on the replacement socket to get the current match state.

#### Client → Server commands
Standard envelope:
```json
{ "command": "READY", "gameId": "...", "matchId": "...", "category": "...", "players": 2, "answer": { "gameId": "...", "questionId": "...", "choice": "A" } }
```
(Only the fields relevant to a given command need to be populated; unused fields are ignored.)

| Command | Required fields | Purpose |
|---|---|---|
| `ADD_TO_WAITING_LIST` | `category` | Join matchmaking. Send `timeOptionId` and `players` to match the selected settings. `displayName` is optional. |
| `LEAVE_QUEUE` | none | Leave the queue and resolve a race with match creation |
| `SYNC_MATCH` | none | Recover the authenticated player's latest match on a new connection |
| `READY` | `gameId` | Signal readiness to start after a match is formed |
| `ANSWER` | `answer.gameId`, `answer.questionId`, `answer.choice` | Submit an answer |
| `GET_CATEGORIES` | none | Reserved command with no response. Read categories from the automatic connection message when available. |

#### Server → Client events
Envelope: `{ "success": bool, "event": "...", "message": "...", "metaData": {...} }`

| Event | When | metaData contents |
|---|---|---|
| `ADDED_TO_WAITING_LIST` | Reply to `ADD_TO_WAITING_LIST` | empty |
| `MATCH_CREATED` | When matchmaking commits a match | `gameId`, `readyUntil` |
| `LEFT_WAITING_LIST` | Reply to `LEAVE_QUEUE` when no match committed | empty |
| `MATCH_SNAPSHOT` | Reply to `SYNC_MATCH` | optional `snapshot` |
| `MATCH_CANCELLED` | When the server's ready deadline expires | `gameId` |
| `QUESTIONS_PUBLISHED` | After all players are ready, then once per round | `gameId`, `roundNumber`, `totalQuestions`, and one entry in `questions[]` with its absolute `ttl` deadline |
| `ROUND_CLOSED` | After everyone answers or the deadline expires | `gameId`, `roundNumber`, and the player's `recap` |
| `ANSWER_ACCEPTED` | Reply to a valid `ANSWER`, sent only to the submitter | `gameId`, `answerFeedback` (`questionId`, `isCorrect`), `leaderBoard` (recomputed standings with answer details omitted) |
| `COMPLETED` | After the final round closes | `leaderBoard` (final). The server then closes the socket. |
| `ERROR` | Various failure conditions (see below) | usually empty |

`ANSWER_ACCEPTED` supplies an immediate verdict for the accepted submission,
before other players answer or the round expires. For example:

```json
{
  "success": true,
  "event": "ANSWER_ACCEPTED",
  "message": "answer accepted",
  "metaData": {
    "gameId": "g1",
    "answerFeedback": {"questionId": "q1", "isCorrect": false},
    "leaderBoard": {"gameId": "g1", "playerPoint": []}
  }
}
```

`isCorrect` compares the submitted answer with the question's correct answer;
it is independent of the points awarded. A false verdict is sent explicitly.
The feedback contains no solution or other players' answers. The full correct
answer remains in the player-specific `ROUND_CLOSED` recap after everyone has
answered or the deadline expires. Rejected submissions receive `ERROR` with no
feedback. Older servers may omit `answerFeedback`; clients must treat its
absence as an acknowledgement without a correctness verdict.

#### Recovering a match

`SYNC_MATCH` uses the authenticated player ID. Use it on a fresh connection before sending gameplay commands. It returns the latest active match or results from the past 24 hours. A successful response with no `snapshot` means there is no recoverable match. An unsuccessful response means the lookup failed and can be retried.

The snapshot has the match ID, settings, and a `phase` of `ready`, `round`, or `finished`. The ready phase has the player's `ready` flag and `readyUntil`. The round phase has the active question and its original deadline. Its optional `submission` has only the authenticated player's accepted choice and verdict. It can also include the previous round's personal recap. The finished phase has final `scores`.

Establish the socket before requesting the snapshot. The game service orders the response with match creation and round transitions. Use the snapshot as the baseline and apply frames received after it. The replacement socket closes the prior connection for the account. The server checks connection ownership while processing `READY` and `ANSWER` so a superseded connection cannot modify a round after the snapshot. Live delivery and ordering use one game-service process, as in Docker Compose. A multi-replica rollout needs shared socket and round coordination.

The default ready window is 60 seconds and can be set with `GAME_service__ready_timeout`. `LEAVE_QUEUE` either confirms the queue exit or returns `MATCH_CREATED` if matchmaking already committed. If a server version predates those commands, a client can continue using its foreground flow and close the socket to leave the queue.

#### Documented WebSocket quirks a client MUST handle defensively (verified in code, not hypothetical)
1. **A failed queue publish can send `ERROR` followed by `ADDED_TO_WAITING_LIST`.** Keep the first failure visible.
2. **`READY` sent before all players are ready produces no immediate success response.** Use the ready timer and the snapshot's readiness flag during recovery.
3. **A duplicate `READY` can return `ERROR` while the server waits for the rest of the roster.**
4. **`GET_CATEGORIES` never responds.**
5. **Only the active question can be answered in a round-based match.** A stale question ID returns `ERROR`.
6. **Round-based matches reject late answers.** The server applies the deadline when it saves the answer.
7. **Reconnect with `SYNC_MATCH`.** The new socket receives current state through the snapshot.
8. **No message-size limits or rate limiting** on the WS layer — not a client concern to work around, but don't assume the server protects itself; the bot's own client library should apply sane limits.
9. **Commands can be sent in any order** with no server-side state machine gate — e.g. sending `ANSWER` before a match even exists just falls through to a generic `{"event":"ERROR","message":"something went wrong"}`. A client-side (bot) state machine should prevent sending commands out of the expected sequence rather than relying on the server to reject them meaningfully.
10. **`ERROR.message` is frequently the generic string `"something went wrong"`** with no machine-readable code — a client cannot branch logic on error type beyond a few specific strings (`"invalid category"`, `"answered to question quickly"`, `"player <id> already answered this question <qid>"`, `"invalid command"`).

Match recovery and standings use the game WebSocket.

---

## 5. QUESTION SERVICE — `/question-service/api/v1/question/...`

### `GET /question-service/api/v1/question/health-check`

Public health check in Compose.

### `GET /question-service/api/v1/question/time-options`

An authenticated request returns an array of enabled question timers. Each option has `id`, `label`, and `seconds`. Use the selected `id` as `timeOptionId` when joining the game socket queue. [Timer handler](../../services/question_app/delivery/http/time_options.go)

### `POST /question-service/api/v1/question/add` — **stub, non-functional**
Both request and response are empty structs (`{}`); the handler performs no validation, no persistence, and always returns `200 {}`. There is no working way to add questions to the bank via any API today — questions must be seeded directly into PostgreSQL.

**Client feature**: None — and even if this service were exposed via the gateway, this endpoint could not support an "admin adds a question" bot feature without backend work.

---

## 6. Kafka Events (not directly client-facing, but shape what the WebSocket eventually delivers)

| Topic | Producer → Consumer | Client relevance |
|---|---|---|
| `GAME_V1_JOIN_MATCH_QUEUE_REQUESTED` | game → match | Indirect: this is how a WS `ADD_TO_WAITING_LIST` command actually lands in the matchmaking queue. |
| `matchMaking_v1_matchUsers` | match → question, game | Indirect: gates when a match/game is created and when questions get generated. |
| `question_v1_questions` | question → game | Indirect: last hop before `QUESTIONS_PUBLISHED` can be sent over WS. |

A client never touches Kafka directly; it's documented here only so a client developer understands the ~15-second scheduler latency and multi-hop pipeline between "join queue" and "match found" (see `client-business-flows.md` for timing).

---

## 7. Missing / Duplicate / Problematic Endpoints for Client Integration

### Missing entirely
- **Token refresh exchange endpoint** (`refresh token → new access token`). The auth-service can *mint* a refresh token, but nothing consumes one to issue a fresh access token. A client has no way to avoid forcing re-login after 24h without this.
- **Logout / token revocation.** JWTs are stateless with no blacklist; a bot "logout" command can only forget the token locally.
- **Get current game/match state** (HTTP, for a bot that reconnects after a backend restart or user restarts the chat). Nothing lets a client ask "what game am I in right now?" outside of the ambiguous WS reconnect behavior in §4.
- **Get game history / past results.** No retrieval API — `docs/business/03-feature-map.md` confirms this was never built.
- **Leave matchmaking queue.** Only joining is supported; no "cancel" command exists.
- **Queue/queue-depth visibility.** No way to show a user "N players waiting in SPORT."
- **Admin question management** — see §5, stub only.
- **Question service gateway route** — service exists and works internally but is unreachable externally; blocks any bot feature that would need to browse/manage questions.

### Duplicate / redundant
- **Joining matchmaking exists twice**: WS `ADD_TO_WAITING_LIST` command and HTTP `POST /match-service/api/v1/addToWaitingList` do the same Redis write, but only the WS path lets the client subsequently receive `MATCH_CREATED`. The HTTP path is effectively dead for a Telegram bot client and should be ignored in the architecture (use WS exclusively).

### Difficult for client integration (design smells to work around, not fix)
- Inconsistent error envelopes: most endpoints return `{"message":"...","error":"..."}`, but `user-service` profile and `match-service` X-User-ID/category errors return **bare JSON strings**, not objects. A client's HTTP layer needs to defensively parse both shapes.
- `match-service`'s `timeout` field is nanoseconds as a raw number, inconsistent with human-readable durations elsewhere.
- `ANSWER_ACCEPTED` echoes `metaData.gameId` and `metaData.answerFeedback.questionId` so the client can match a verdict to its pending submission and ignore stale or duplicate verdicts. Older servers that omit feedback still lack this question correlation.
- No structured/machine-readable error codes over WebSocket (see §4, point 10) — client must pattern-match on message strings, which is brittle.

---

## 8. Recommended Improvements (summary — see `backend-improvements.md` for full detail)

1. Add a token refresh endpoint (`user-service` or `auth-service`).
2. Standardize error response bodies to the `{"message","error","code"}` shape and add machine-readable WebSocket error codes.
3. Route question timers through the Kubernetes ingress for Android staging deployments that use those manifests.
4. Add acknowledgments for `READY` before the roster is ready and for `GET_CATEGORIES`.
