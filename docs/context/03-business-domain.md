# 03 — Business Domain

## Business Concepts

BrainBlitz is a **live quiz competition** platform. Players are paired against strangers in real time, answer time-limited questions, and see a live leaderboard at the end.

## Domain Terminology

| Term | Definition |
|---|---|
| **Category** | A topic area for quiz questions. Supported values: `SPORT`, `MUSIC`, `TECH` |
| **Difficulty** | Question difficulty level: `EASY`, `MEDIUM`, `HARD` |
| **Waiting List** | A per-category Redis sorted set of players waiting to be matched |
| **Match** | A grouping of 2+ players assigned the same questions to compete against each other. Identified by a ULID-based `matchId` |
| **Game** | A single play session derived from a match. Stored in MongoDB as a document. Identified by a MongoDB ObjectID (`gameId`) |
| **Game Status** | Lifecycle state of a game session: `UNKNOWN` → `INITIALIZED` → `PENDING` → `CREATED` → `STARTED` → `FINISHED` |
| **Question** | A quiz question with content, 4 choices, one correct answer, category, and difficulty |
| **Player Answer** | A record of a player's choice for a specific question within a game, including timing data |
| **Leaderboard** | Ranked list of players by total points within a game |
| **Valid Answer Time** | The deadline by which a player must answer a given question to score points |
| **Time Bonus** | Additional points awarded for answering quickly before the deadline |
| **Access Token** | Short-lived JWT (default 24h) used for API authentication |
| **Refresh Token** | Longer-lived JWT (default 120h) used to obtain new access tokens |
| **Role** | User permission level: `user` or `admin` |
| **Display Name** | Auto-generated from the email username part (e.g., `alice@example.com` → `alice`) |

## Entities

### User

```
id           SERIAL PRIMARY KEY
username     VARCHAR(100) UNIQUE   (stores email)
display_name VARCHAR(100)
role         VARCHAR(10)           ('user' or 'admin')
password     VARCHAR(255)          (bcrypt hash)
created_at   TIMESTAMPTZ
updated_at   TIMESTAMPTZ
```

### Question (PostgreSQL — question_db)

```
id             UUID PRIMARY KEY
content        TEXT
correct_answer TEXT
choices        TEXT[]              (array of 4 choices)
category       TEXT                ('SPORT', 'MUSIC', 'TECH')
difficulty     TEXT                ('EASY', 'MEDIUM', 'HARD')
created_at     TIMESTAMP
```

### user_question_history (PostgreSQL — question_db)

```
user_id     BIGINT
question_id UUID
seen_at     TIMESTAMP
PRIMARY KEY (user_id, question_id)
```

Tracks which questions a user has seen. Questions seen within 30 days are excluded from new matches via `GetProperQuestions`.

### match_questions (PostgreSQL — question_db)

```
match_id    UUID
question_id UUID
PRIMARY KEY (match_id, question_id)
```

> Note: This table is defined in migrations but `GetRandomQuestions` does not use it. Its intended use is to persist the question assignment per match, but the current code path uses Kafka to deliver questions and Redis/MongoDB for runtime storage.

### Game (MongoDB — collection: `game`)

```json
{
  "_id":       ObjectID,
  "players":   [uint64, ...],    // user IDs
  "match_id":  string,
  "category":  [string, ...],
  "status":    string,
  "questions": [...],            // embedded Question array
  "created_at": datetime,
  "updated_at": datetime
}
```

### PlayerAnswer (MongoDB — collection: `player_answers`)

```json
{
  "game_id":             string,
  "question_id":         string,
  "player_id":           string,
  "player_choice":       string,
  "correct_choice":      string,
  "answer_time":         datetime,
  "valid_time_to_answer": datetime,
  "time_diff":           duration,
  "Option":              [string, ...],
  "point":               int,
  "category":            string
}
```

### WaitingMember (Redis — sorted set per category)

Key: `waiting_users:<CATEGORY>` (configurable prefix)
Score: Unix microseconds timestamp of when the user joined

## Core Workflows

### 1. User Registration

**Trigger**: POST `/user-service/public/api/v1/signup`

1. User provides `email` and `password`
2. Service validates email format (`pkg/email.IsValid`)
3. Service validates password is non-empty
4. Password is hashed with bcrypt
5. `display_name` is derived from the email local part
6. User record is inserted into PostgreSQL with role `user`
7. Response: `{ displayName: string }`

**Failure cases**: duplicate email (409 equivalent, 400 with duplicate message), invalid email format.

### 2. User Login

**Trigger**: POST `/user-service/public/api/v1/login`

1. User provides `email` and `password`
2. User record is looked up by email in PostgreSQL
3. Password is checked against bcrypt hash
4. If correct: User Service calls Auth Service via gRPC to create access and refresh tokens (embedding `id` and `role` claims)
5. Response: `{ id, accessToken, refreshToken }`

**Failure cases**: user not found (mapped to forbidden), wrong password (forbidden).

### 3. Token Validation (Traefik middleware)

**Trigger**: Every protected HTTP request via Traefik ForwardAuth

1. Traefik forwards the request to `POST /auth-service/api/v1/validate-token` with the `Authorization` header
2. Auth Service parses and validates the JWT signature and expiry
3. On success: extracts `id` and `role` claims from the token
4. Auth Service sets response headers: `X-User-ID`, `X-User-Role`, `X-Auth-Data`
5. Traefik propagates these headers to the downstream service

### 4. Matchmaking Flow

**Trigger**: Client sends `ADD_TO_WAITING_LIST` command over WebSocket

1. Game Service receives `{ command: "ADD_TO_WAITING_LIST", category: "SPORT" }` over WebSocket
2. Sets user's game status to `INITIALIZED` in Redis
3. Publishes `GAME_V1_JOIN_MATCH_QUEUE_REQUESTED` to Kafka (proto-encoded: `userId`, `category`)
4. Match Service consumer receives the event and stores the user in a Redis sorted set keyed by `waiting_users:<CATEGORY>` with score = current Unix microseconds

**Match Scheduler (every 15 seconds)**:

1. Queries all waiting lists by category (ZRange with time window)
2. Sorts members by timestamp (FIFO)
3. Groups by category; requires at least 2 players per category
4. If an odd number, excludes the most recent joiner
5. Generates a ULID-based `matchId` for each pair/group
6. Publishes `matchMaking_v1_matchUsers` to Kafka
7. Removes matched members from Redis sorted sets

### 5. Game Creation

**Trigger**: Match Service publishes `matchMaking_v1_matchUsers`

Both Question Service and Game Service consume this event in parallel:

**Question Service path**:
1. Fetches 10 random questions of difficulty `EASY` for the matched category from PostgreSQL
2. Publishes `question_v1_questions` to Kafka (proto-encoded: `matchId`, `[]Question`)

**Game Service path**:
1. Creates a game document in MongoDB with status `PENDING`
2. Saves users' game status to `PENDING` in Redis
3. Notifies all connected players via WebSocket: `{ event: "MATCH_CREATED", gameId: "..." }`

**Game Service also consumes `question_v1_questions`**:
1. Stores questions in MongoDB under the game document (by `matchId`)
2. Caches them in Redis keyed by `game_questions_<gameId>`

### 6. Game Start (READY phase)

**Trigger**: Client sends `{ command: "READY", gameId: "..." }` over WebSocket

1. Game Service increments the "ready player" count in Redis for the game
2. When all expected players are ready:
   - Sets `ValidAnswerTime` for each question: question `i` gets deadline `now + (validAnswerTimeout * (i+1))`
   - Sends all questions to all players via WebSocket: `{ event: "QUESTIONS_PUBLISHED", ... }`
   - Schedules an Asynq task `"game:completed"` with `ProcessIn = total TTL`

### 7. Answering Questions

**Trigger**: Client sends `{ command: "ANSWER", answer: { gameId, questionId, choice } }` over WebSocket

1. Game Service retrieves questions from Redis
2. Validates: the answer must not be received before the question's `ValidAnswerTime` (catches cheating by submitting too early)
   - Note: The validation logic checks `if validTime.Sub(answerTime) > totalValidAnswerTimeout`, treating this as "answered too quickly" which is actually the correct window check
3. Checks for duplicate answers (same player, same question in MongoDB)
4. Computes score via `CalculateScore`:
   - `0` if incorrect or after deadline
   - `BaseScore + bonus` if correct and on time
   - Bonus = `MaxBonus * (timeDiff / BonusDeadline)`, where `timeDiff = validAnswerTime - answerTime`
5. Inserts `PlayerAnswer` to MongoDB
6. Returns current leaderboard to the answering player

### 8. Game Completion

**Trigger**: Asynq task `"game:completed"` fires after total TTL

1. Game Service fetches the full leaderboard from MongoDB
2. Aggregates all player answers into ranked `PlayerPoint` objects
3. Sends `{ event: "COMPLETED", leaderBoard: {...} }` to all players via WebSocket
4. Closes all WebSocket connections for the game

## Business Rules

1. A match requires **at least 2 players** in the same category.
2. A player can only answer a question **once** per game.
3. An answer submitted **before** the question's valid answer window is rejected (anti-cheat).
4. An answer submitted **after** the deadline scores 0 (but is still recorded).
5. Categories are fixed at compile time: `SPORT`, `MUSIC`, `TECH`.
6. Questions are selected with difficulty `EASY` only (hardcoded in `ConsumeMatchCreated`).
7. The number of questions per match is fixed at **10** (hardcoded `limit = 10`).
8. Player count per match is **2** (visible in `getCategories` which returns `users: [2]`).
9. Waiting list selection window is `-20m` to `now` (players who joined more than 20 minutes ago are excluded).
10. User roles are `user` and `admin`; currently all registered users get the `user` role.

## WebSocket Protocol

The Game Service WebSocket endpoint (`GET /game-service/api/v1/process-game`) accepts JSON command frames and sends JSON event frames.

### Client → Server (Commands)

```json
{ "command": "ADD_TO_WAITING_LIST", "category": "SPORT" }
{ "command": "READY", "gameId": "<gameId>" }
{ "command": "ANSWER", "gameId": "<gameId>", "answer": { "gameId": "<gameId>", "questionId": "<id>", "choice": "A" } }
{ "command": "GET_CATEGORIES" }
```

### Server → Client (Events)

| Event | When | Payload fields |
|---|---|---|
| `ADDED_TO_WAITING_LIST` | After join queue accepted | `success`, `message` |
| `MATCH_CREATED` | Matchmaking complete | `metaData.gameId` |
| `QUESTIONS_PUBLISHED` | All players ready | `metaData.questions[]` |
| `ANSWER_ACCEPTED` | Answer stored | `metaData.leaderBoard` |
| `COMPLETED` | Game time expired | `metaData.leaderBoard` |
| `ERROR` | Invalid command / internal error | `message` |

On first connection when user status is `UNKNOWN`, the server also sends:
```json
{ "categories": ["SPORT","MUSIC","TECH"], "numberOfPlayers": [2] }
```

## Assumptions

- The `question_db.users` table defined in Question Service migrations is a denormalized copy for future use; currently it is not populated by the application.
- The `match_questions` table is defined but unused in the current code path.
- `GetProperQuestions` (history-aware question selection) exists but is not called by `ConsumeMatchCreated` — `GetRandomQuestions` is used instead.
