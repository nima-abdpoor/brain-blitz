# Refactoring Overview

## Current Code Quality Assessment

### Maintainability: 6 / 10

The codebase has a clear, consistent layered structure (`delivery → service → repository`) that is easy to navigate. However, several issues lower maintainability:
- Business logic assumptions are not documented in code
- Configuration is partially duplicated across services (each service redeclares `Category`, `Difficulty` types)
- Kafka topic strings are scattered (only one is a constant; others are inline strings)
- No application-level metrics, making production debugging difficult

### Architecture Score: 6 / 10

The microservice decomposition is sound and the service boundaries make sense. Points deducted for:
- In-memory WebSocket connection map prevents horizontal scaling of Game Service
- Kafka consumer uses low-level partition consumer (not consumer groups) — blocks horizontal scaling
- No distributed cache or pub/sub for WebSocket fan-out across multiple Game Service instances
- Auth adapter bug (refresh token returns access token) undermines security model

### Complexity Hotspots

| File | Lines | Complexity | Issue |
|---|---|---|---|
| `services/game_app/service/service.go` | 689 | High | Monolithic; handles WS protocol, game state, match events, question events, task scheduling coordination |
| `services/game_app/repository/game.go` | 515 | High | Mixed MongoDB + Redis logic; scoring algorithm; question timing logic |
| `services/match_app/service/service.go` | 200 | Medium | Matchmaking algorithm is readable but has subtle edge cases |
| `services/auth_app/service/service.go` | 143 | Low | Clean but contains debug `fmt.Println` |

### Testing Coverage

Limited. Unit tests exist for:
- Auth service token logic
- Match service validator
- User service basic logic
- User repository (sqlmock)

Missing tests for:
- Game service (all WebSocket protocol logic)
- Question service
- All adapter implementations
- Repository integration tests (Game, Match)

### Technical Debt Score: Medium-High

Approximately 15 `// todo` comments, 2 confirmed bugs, and multiple unimplemented features (AddQuestion, game status update on completion, metrics, history-aware question selection).
