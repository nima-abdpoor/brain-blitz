# Business Domain Model

## Actors

| Actor | Description |
|---|---|
| **Player** | Registered user who participates in quiz games. Role: `user` |
| **Admin** | Privileged user. Role: `admin`. No additional capabilities implemented yet (assumption: future administration of questions, categories, etc.) |
| **System** | Internal automated processes: Match Scheduler, Asynq task worker, Kafka consumers |

## Core Entities

```
┌─────────────┐         ┌─────────────────┐         ┌────────────────┐
│    User     │─────────│      Game       │─────────│  PlayerAnswer  │
│             │  plays  │                 │ contains│                │
│ id          │    n    │ id (ObjectID)   │   n     │ game_id        │
│ username    │         │ players[]       │         │ player_id      │
│ password    │         │ match_id        │         │ question_id    │
│ display_name│         │ category[]      │         │ player_choice  │
│ role        │         │ status          │         │ correct_choice │
│ created_at  │         │ questions[]     │         │ answer_time    │
│ updated_at  │         │ created_at      │         │ point          │
└─────────────┘         └─────────────────┘         └────────────────┘
                               │ 1
                               │
                        ┌──────┴──────┐
                        │    Match    │
                        │             │
                        │ matchId     │
                        │ category    │
                        │ userIds[]   │
                        └─────────────┘
                               │
              ┌────────────────┘
              │
    ┌─────────┴──────────┐
    │     Question       │
    │                    │
    │ id (UUID)          │
    │ content            │
    │ correct_answer     │
    │ choices TEXT[]     │
    │ category           │
    │ difficulty         │
    └────────────────────┘
```

## Entity Relationships

| Relationship | Cardinality | Details |
|---|---|---|
| User plays Games | M:N | Via `game.players[]` array of user IDs |
| Game has Questions | 1:N | Embedded array in Game document + linked via matchId |
| Game produces PlayerAnswers | 1:N | `player_answers.game_id` foreign key |
| Player submits PlayerAnswers | 1:N | `player_answers.player_id` |
| Match contains Users | 1:N | `AllMatchedUsers` proto payload |
| Question belongs to Category | N:1 | `questions.category` field |

## Terminology

| Term | Type | Values / Notes |
|---|---|---|
| `Category` | Enum | `SPORT`, `MUSIC`, `TECH`, `GEOGRAPHY`, `SCIENCE`, `HISTORY` |
| `Difficulty` | Enum | `EASY`, `MEDIUM`, `HARD` |
| `GameStatus` | Enum | `UNKNOWN`, `INITIALIZED`, `PENDING`, `CREATED`, `STARTED`, `FINISHED` |
| `Role` | Enum | `user` (1), `admin` (2) |
| `MatchId` | String | ULID format — time-sortable, collision-resistant |
| `GameId` | String | MongoDB ObjectID hex string |
| `ValidAnswerTime` | Timestamp | Per-question deadline; question `i` gets `now + (timeout * (i+1))` |
| `BaseScore` | Int | Configurable, default 5 |
| `MaxBonus` | Int | Configurable, default 10 |
| `BonusDeadline` | Duration | Configurable, default 115s |
