# Use Cases

## UC-01: Register New Player

**Actors**: New User

**Preconditions**: User has an email address; server is reachable.

**Main Flow**:
1. User submits `POST /user-service/public/api/v1/signup` with `{ email, password }`
2. System validates email format
3. System validates password is non-empty
4. System hashes password with bcrypt
5. System derives `displayName` from email local part
6. System inserts user into PostgreSQL with role `user`
7. System returns `{ displayName }`

**Alternative Flow — Duplicate Email**:
- At step 6, if email already exists: system returns 400 with duplicate username message

**Alternative Flow — Invalid Email**:
- At step 2: system returns 400 error

**Postconditions**: User account exists in PostgreSQL; user can now log in.

---

## UC-02: Login

**Actors**: Registered Player

**Preconditions**: User account exists.

**Main Flow**:
1. User submits `POST /user-service/public/api/v1/login` with `{ email, password }`
2. System validates email and password are non-empty
3. System fetches user record by email from PostgreSQL
4. System verifies password against bcrypt hash
5. System calls Auth Service gRPC to create access token (claims: `id`, `role`)
6. System calls Auth Service gRPC to create refresh token (same claims)
7. System returns `{ id, accessToken, refreshToken }`

**Alternative Flow — Wrong Password**:
- At step 4: system returns 403

**Alternative Flow — User Not Found**:
- At step 3: system returns 403

**Postconditions**: Client holds a valid access token for 24 hours and a refresh token for 120 hours.

---

## UC-03: Play a Quiz Game

**Actors**: Authenticated Player (requires a valid access token)

**Preconditions**: Player is logged in; WebSocket client is available; at least one other player must join the same category within 20 minutes.

**Main Flow**:
1. Player opens WebSocket connection to `GET /game-service/api/v1/process-game` with `Authorization: Bearer <token>`
2. Traefik validates token via ForwardAuth; injects `X-User-ID`
3. Game Service upgrades connection; checks player's current game status
4. If status is UNKNOWN: Game Service sends available categories and player count options
5. Player sends `{ command: "ADD_TO_WAITING_LIST", category: "SPORT" }`
6. Game Service publishes join event to Kafka; Match Service stores player in Redis queue
7. Game Service responds with `{ event: "ADDED_TO_WAITING_LIST" }`
8. Player waits; within 15 seconds, Match Service scheduler pairs players
9. Match Service publishes match event; Game Service creates game in MongoDB; player receives `{ event: "MATCH_CREATED", gameId: "..." }`
10. Question Service publishes questions; Game Service stores them
11. Player sends `{ command: "READY", gameId: "..." }`
12. When all players are ready, Game Service sends all questions: `{ event: "QUESTIONS_PUBLISHED", questions: [...] }`
13. Player submits answers: `{ command: "ANSWER", answer: { gameId, questionId, choice } }`
14. For each answer: Game Service validates timing, records answer, returns current leaderboard
15. After all question deadlines expire, Asynq task fires, Game Service sends `{ event: "COMPLETED", leaderBoard: [...] }` to all players
16. Game Service closes all WebSocket connections

**Alternative Flows**:
- Invalid category at step 5 → `{ event: "ERROR", message: "invalid category" }`
- Answer too early → `{ event: "ERROR" }` (anti-cheat rejection)
- Answer after deadline → recorded with 0 points
- Not enough players → player waits indefinitely (no timeout currently)

**Postconditions**: Game document in MongoDB has all answers; player has seen their final score.

---

## UC-04: View Player Profile

**Actors**: Authenticated Player

**Preconditions**: Player is logged in.

**Main Flow**:
1. Client calls `GET /user-service/api/v1/profile` with `Authorization: Bearer <token>`
2. Traefik validates token; injects `X-User-ID`
3. User Service fetches user by ID from PostgreSQL
4. Returns `{ id, username, displayName, role, createdAt, updatedAt }`

**Alternative Flow — User Not Found**:
- At step 3: returns 404

**Postconditions**: None (read-only).

---

## UC-05: Add Player to Matchmaking Queue (HTTP fallback)

**Actors**: Authenticated Player

**Preconditions**: Player is logged in.

**Note**: This is a direct HTTP alternative to the WebSocket flow for joining the queue. It does not establish a persistent connection, so the player would not receive `MATCH_CREATED` notifications unless they also have a WebSocket connection.

**Main Flow**:
1. Client calls `POST /match-service/api/v1/addToWaitingList` with `{ category }` and `Authorization: Bearer <token>`
2. Traefik validates token; injects `X-User-ID`
3. Match Service validates category
4. Match Service adds player to Redis sorted set for the category
5. Returns `{ timeout: "20m" }`

**Postconditions**: Player is in Redis waiting list; will be matched by next scheduler run.

---

## UC-06: Token Validation (System Internal)

**Actors**: Traefik (System)

**Preconditions**: Incoming request has `Authorization` header.

**Main Flow**:
1. Traefik extracts `Authorization` header
2. Forwards request to `Auth Service /api/v1/validate-token`
3. Auth Service parses JWT; verifies signature with `SECRET_KEY`; checks expiry
4. Extracts `id` and `role` claims
5. Sets response headers: `X-User-ID`, `X-User-Role`, `X-Auth-Data`
6. Traefik injects headers into original request and forwards to target service

**Alternative Flow — Invalid Token**:
- At step 3: Auth Service returns `{ valid: false }` → Traefik returns 401 to client

**Postconditions**: Downstream service has access to caller's identity without decoding JWT itself.
