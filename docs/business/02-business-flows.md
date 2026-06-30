# Business Flows

## Flow 1: User Registration

**Purpose**: Allow a new player to create an account.

**Trigger**: POST `/user-service/public/api/v1/signup`

**Steps**:
1. Client submits `{ email, password }`
2. Email format validated
3. Password non-empty check
4. Bcrypt hash computed
5. `display_name` derived from email local part (e.g. `alice@example.com` → `alice`)
6. User inserted into PostgreSQL `users` table with role `user`
7. Response: `{ displayName }`

**Decision points**:
- Invalid email → 400 error
- Duplicate email → 400 error (contains "duplicate" in message)

**Outputs**: User record in PostgreSQL; `displayName` returned to client.

**Failure cases**: PostgreSQL unavailable → 500 error.

**Dependencies**: PostgreSQL.

---

## Flow 2: User Login

**Purpose**: Authenticate a player and issue tokens.

**Trigger**: POST `/user-service/public/api/v1/login`

**Steps**:
1. Client submits `{ email, password }`
2. User record fetched from PostgreSQL by email
3. bcrypt password check
4. If correct: User Service calls Auth Service via gRPC
   - `GetAccessToken` with claims `{ id: userId, role: userRole }`
   - `GetRefreshToken` with same claims
5. Response: `{ id, accessToken, refreshToken }`

**Decision points**:
- Email not found → 403 error
- Password mismatch → 403 error
- Auth gRPC call fails → 500 error

**Outputs**: Access token (24h), refresh token (120h).

**Failure cases**: Auth Service unreachable → 500; PostgreSQL unavailable → 500.

**Dependencies**: PostgreSQL, Auth Service (gRPC).

---

## Flow 3: API Authentication (Traefik ForwardAuth)

**Purpose**: Protect all non-public routes.

**Trigger**: Every HTTP request to a protected route.

**Steps**:
1. Traefik receives request with `Authorization: Bearer <token>`
2. Traefik calls `POST /auth-service/api/v1/validate-token` forwarding the Authorization header
3. Auth Service parses and verifies JWT signature and expiry
4. On success: extracts `id`, `role` from claims; sets `X-User-ID`, `X-User-Role`, `X-Auth-Data` response headers
5. Traefik forwards original request to target service with injected headers
6. Downstream service reads `X-User-ID` to identify the caller

**Decision points**:
- Missing Authorization header → 401 immediately
- Invalid/expired token → 401
- Valid token → proceed to downstream service

**Outputs**: `X-User-ID`, `X-User-Role`, `X-Auth-Data` headers available in downstream service.

**Failure cases**: Auth Service unreachable → 502 from Traefik.

---

## Flow 4: Join Matchmaking Queue

**Purpose**: A player signals they want to play.

**Trigger**: WebSocket client sends `{ command: "ADD_TO_WAITING_LIST", category: "SPORT" }`

**Steps**:
1. Game Service receives command on established WebSocket connection
2. Validates category is one of `SPORT`, `MUSIC`, `TECH`
3. Sets user's game status to `INITIALIZED` in Redis
4. Publishes `GAME_V1_JOIN_MATCH_QUEUE_REQUESTED` event to Kafka
5. Match Service consumer receives event; adds user to Redis sorted set `waiting_users:SPORT` with score = current Unix microseconds
6. Game Service sends `{ event: "ADDED_TO_WAITING_LIST", success: true }` to client

**Decision points**:
- Invalid category → error response, no Kafka publish
- Kafka publish failure → error response to client

**Outputs**: User in Redis waiting list; client notified.

**Failure cases**: Kafka unavailable → error response.

---

## Flow 5: Matchmaking

**Purpose**: Pair waiting players and create a match.

**Trigger**: Match Service scheduler (every 15 seconds).

**Steps**:
1. Query Redis sorted sets for all categories (time window: last 20 minutes)
2. Sort all waiting members by join timestamp (oldest first)
3. Group by category
4. For each category with ≥2 members: take an even number (drop most recent if odd)
5. Generate a ULID `matchId` for the group
6. Publish `matchMaking_v1_matchUsers` to Kafka with `{ matchId, userIds[], category[] }`
7. Remove matched members from Redis sorted sets

**Decision points**:
- Category has fewer than 2 members → skip that category
- Odd number of members → drop the most recently joined member

**Outputs**: `matchMaking_v1_matchUsers` Kafka event; matched users removed from waiting lists.

**Failure cases**: Redis unavailable → scheduler logs error, retries on next tick. Kafka publish fails → users remain in waiting list but receive no match.

---

## Flow 6: Game Session Creation

**Purpose**: Set up a game session after a match is formed.

**Trigger**: Both Game Service and Question Service consume `matchMaking_v1_matchUsers`.

**Parallel paths**:

**Question Service**:
1. Fetches 10 random questions of difficulty EASY matching the match's category from PostgreSQL
2. Publishes `question_v1_questions` to Kafka

**Game Service**:
1. Creates a `game` document in MongoDB with status `PENDING`, embeds player IDs and matchId
2. Updates each player's game status to `PENDING` in Redis
3. Pushes `{ event: "MATCH_CREATED", gameId }` to all connected players via WebSocket

**Game Service** (also consumes `question_v1_questions`):
1. Saves questions to MongoDB game document (update by matchId)
2. Caches questions in Redis under `game_questions_<gameId>`

**Outputs**: Game document in MongoDB; questions in Redis cache; players notified of their game ID.

---

## Flow 7: Game Start (Ready Phase)

**Purpose**: Synchronize all players before delivering questions.

**Trigger**: Each player sends `{ command: "READY", gameId: "<gameId>" }`.

**Steps**:
1. Game Service increments the ready player count in Redis
2. When ready count reaches expected player count:
   a. Sets `ValidAnswerTime` for each question (staggered deadlines)
   b. Sends all questions to all players: `{ event: "QUESTIONS_PUBLISHED", questions: [...] }`
   c. Enqueues Asynq task `"game:completed"` with delay = total TTL

**Decision points**:
- Not all players ready yet → wait; no action
- All players ready → start game immediately

**Outputs**: Players receive question set; countdown timer implicitly starts per `ValidAnswerTime` values.

---

## Flow 8: Answering a Question

**Purpose**: Record a player's answer and provide immediate feedback.

**Trigger**: Player sends `{ command: "ANSWER", answer: { gameId, questionId, choice } }`.

**Steps**:
1. Load question data from Redis
2. Check question `ValidAnswerTime` — answer must arrive after the answer-time window opens (anti-cheat: too-early submission rejected)
3. Check for duplicate answer (player already answered this question in MongoDB)
4. Compute score: `CalculateScore(isCorrect, answerTime, validAnswerTime)`
5. Insert `PlayerAnswer` record to MongoDB
6. Fetch and return current leaderboard to answering player

**Decision points**:
- Answer arrives too early → 400 error "answered to question quickly"
- Duplicate answer → error (player cannot change their answer)
- Answer after deadline → score = 0, but still recorded

**Outputs**: `PlayerAnswer` in MongoDB; leaderboard returned to player.

---

## Flow 9: Game Completion

**Purpose**: End the game and deliver final results.

**Trigger**: Asynq task `"game:completed"` fires after total question TTL.

**Steps**:
1. Fetch all `PlayerAnswer` records for the game from MongoDB
2. Aggregate points by player; sort descending
3. Send `{ event: "COMPLETED", leaderBoard: [...] }` to all players via WebSocket
4. Close all WebSocket connections for the game

**Outputs**: Final leaderboard delivered to all players; connections closed.

**Failure cases**: If leaderboard query fails, players receive an error event.

---

## Flow 10: User Profile Retrieval

**Purpose**: Return a player's profile.

**Trigger**: GET `/user-service/api/v1/profile` (requires authentication).

**Steps**:
1. Traefik validates JWT via ForwardAuth; injects `X-User-ID`
2. User Service reads `X-User-ID` header
3. Fetches user record from PostgreSQL by ID
4. Returns `{ id, username, displayName, role, createdAt, updatedAt }`

**Failure cases**: User not found → 404.
