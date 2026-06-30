# Architecture Diagrams

## System Architecture Overview

```mermaid
graph TB
    Client["Client\n(Browser / WebSocket)"]
    
    subgraph Gateway["API Gateway"]
        Traefik["Traefik v2.10\n:80 HTTP, :8080 Dashboard"]
    end
    
    subgraph Services["Microservices"]
        AuthSvc["Auth Service\n:5000 HTTP | :6000 gRPC"]
        UserSvc["User Service\n:5001 HTTP | :6001 gRPC"]
        MatchSvc["Match Service\n:5002 HTTP"]
        GameSvc["Game Service\n:5003 HTTP/WS"]
        QuestionSvc["Question Service\n:5004 HTTP"]
    end
    
    subgraph Infra["Infrastructure"]
        Kafka["Kafka\n:9092"]
        Redis["Redis\n:6379"]
        PG["PostgreSQL\n:5433"]
        Mongo["MongoDB Replica Set\nrs0 (3 nodes)"]
    end
    
    subgraph Observability["Observability"]
        Prometheus["Prometheus\n:9090"]
        Grafana["Grafana\n:3000"]
    end

    Client -->|"HTTP/WS"| Traefik
    Traefik -->|"ForwardAuth"| AuthSvc
    Traefik -->|"route"| UserSvc
    Traefik -->|"route"| MatchSvc
    Traefik -->|"route"| GameSvc
    
    UserSvc -->|"gRPC GetAccessToken"| AuthSvc
    UserSvc --> PG
    
    MatchSvc -->|"ZAdd/ZRange/ZRem"| Redis
    MatchSvc -->|"Publish/Consume"| Kafka
    
    GameSvc -->|"game state"| Redis
    GameSvc -->|"game docs"| Mongo
    GameSvc -->|"Publish/Consume"| Kafka
    GameSvc -->|"Asynq tasks"| Redis
    
    QuestionSvc --> PG
    QuestionSvc -->|"Publish/Consume"| Kafka
    
    Prometheus -->|"scrape :9216"| Mongo
    Grafana --> Prometheus
```

## Kafka Event Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant GS as Game Service
    participant MS as Match Service
    participant QS as Question Service
    participant K as Kafka

    C->>GS: WS: ADD_TO_WAITING_LIST (category=SPORT)
    GS->>K: Publish GAME_V1_JOIN_MATCH_QUEUE_REQUESTED
    K-->>MS: Consume → AddToWaitingList (Redis)
    
    Note over MS: Scheduler every 15s
    MS->>MS: MatchWaitUsers()
    MS->>K: Publish matchMaking_v1_matchUsers
    
    par Question Service path
        K-->>QS: Consume matchMaking_v1_matchUsers
        QS->>QS: GetRandomQuestions (PostgreSQL)
        QS->>K: Publish question_v1_questions
    and Game Service path
        K-->>GS: Consume matchMaking_v1_matchUsers
        GS->>GS: CreateGame (MongoDB)
        GS->>C: WS: MATCH_CREATED {gameId}
    end
    
    K-->>GS: Consume question_v1_questions
    GS->>GS: SaveQuestions (MongoDB + Redis)
    
    C->>GS: WS: READY {gameId}
    GS->>GS: SetValidAnswerTimes (Redis)
    GS->>C: WS: QUESTIONS_PUBLISHED {questions[]}
    GS->>GS: Schedule Asynq task game:completed
    
    C->>GS: WS: ANSWER {questionId, choice}
    GS->>GS: SavePlayerAnswer (MongoDB)
    GS->>C: WS: ANSWER_ACCEPTED {leaderBoard}
    
    Note over GS: Asynq fires after total TTL
    GS->>GS: GetLeaderBoard (MongoDB)
    GS->>C: WS: COMPLETED {leaderBoard}
```

## Request Lifecycle — Protected Route

```mermaid
sequenceDiagram
    participant C as Client
    participant T as Traefik
    participant A as Auth Service
    participant S as Target Service
    participant DB as Database

    C->>T: GET /user-service/api/v1/profile<br/>Authorization: Bearer <token>
    T->>A: POST /api/v1/validate-token<br/>Authorization: Bearer <token>
    A->>A: Parse JWT, verify signature
    A-->>T: 200 OK<br/>X-User-ID: 42<br/>X-User-Role: user
    T->>S: GET /api/v1/profile<br/>X-User-ID: 42
    S->>DB: SELECT * FROM users WHERE id=42
    DB-->>S: User record
    S-->>T: 200 {id, username, displayName, ...}
    T-->>C: 200 {id, username, displayName, ...}
```

## Service Startup Flow

```mermaid
flowchart TD
    A[main] --> B[cfgloader.Load YAML + env vars]
    B --> C[logger.Init]
    C --> D{Service type?}
    
    D -->|user, question| E[postgresql.Connect]
    D -->|game| F[mongo.NewDB]
    D -->|match, game, question| G[broker.NewKafkaBroker]
    D -->|user| H[grpc.NewClient → Auth Service]
    
    E --> I[postgresqlmigrator.Up]
    I --> J[Service.Setup]
    F --> J
    G --> J
    H --> J
    
    J --> K[app.Start]
    K --> L[signal.NotifyContext SIGINT/SIGTERM]
    L --> M[startServers goroutines]
    
    M --> N[HTTP Server]
    M --> O[gRPC Server]
    M --> P[Kafka Consumer]
    M --> Q[Scheduler]
    M --> R[Asynq Worker]
    
    L --> S{Signal received?}
    S -->|yes| T[shutdownServers with timeout]
    T --> U[wg.Wait]
    U --> V[Exit]
```

## Game State Machine

```mermaid
stateDiagram-v2
    [*] --> UNKNOWN : Player connects via WS
    UNKNOWN --> INITIALIZED : ADD_TO_WAITING_LIST command
    INITIALIZED --> PENDING : Match Service pairs players
    PENDING --> PENDING : READY received (not all ready)
    PENDING --> STARTED : All players READY
    STARTED --> FINISHED : Asynq game:completed fires
    FINISHED --> [*]

    note right of UNKNOWN : Server sends categories list
    note right of INITIALIZED : User added to Redis waiting list
    note right of PENDING : Game created in MongoDB;\nQuestions cached in Redis
    note right of STARTED : Questions sent to all players;\nAnswer window open
    note right of FINISHED : Final leaderboard sent;\nWS connections closed
```

## Package Dependency Graph

```mermaid
graph LR
    subgraph cmd
        auth_main["cmd/auth"]
        user_main["cmd/user"]
        match_main["cmd/match"]
        game_main["cmd/game"]
        question_main["cmd/question"]
    end

    subgraph services
        auth_app["services/auth_app"]
        user_app["services/user_app"]
        match_app["services/match_app"]
        game_app["services/game_app"]
        question_app["services/question_app"]
    end

    subgraph adapters
        broker_a["adapter/broker"]
        redis_a["adapter/redis"]
        ws_a["adapter/websocket"]
        tq_a["adapter/task-queue"]
        auth_a["adapter/auth"]
    end

    subgraph pkg
        grpc_p["pkg/grpc"]
        http_p["pkg/http_server"]
        logger_p["pkg/logger"]
        err_p["pkg/err_app"]
        pg_p["pkg/postgresql"]
        mongo_p["pkg/mongo"]
        cfg_p["pkg/cfg_loader"]
        cache_p["pkg/cache_manager"]
    end

    subgraph contract
        auth_c["contract/auth/golang"]
        match_c["contract/match/golang"]
        question_c["contract/question/golang"]
        event_c["contract/event"]
    end

    auth_main --> auth_app
    user_main --> user_app
    match_main --> match_app
    game_main --> game_app
    question_main --> question_app

    auth_app --> grpc_p & http_p & logger_p & err_p & auth_c
    user_app --> auth_a & redis_a & pg_p & grpc_p & cache_p
    match_app --> broker_a & redis_a & http_p & match_c & event_c
    game_app --> broker_a & redis_a & ws_a & tq_a & mongo_p & match_c & question_c & event_c
    question_app --> broker_a & pg_p & match_c & question_c

    auth_a --> auth_c
```

## Scoring Algorithm

```mermaid
flowchart TD
    A[Player submits ANSWER] --> B{Is correct?}
    B -->|No| Z[score = 0]
    B -->|Yes| C{After validAnswerTime?}
    C -->|Yes| Z
    C -->|No| D[timeDiff = validAnswerTime - answerTime]
    D --> E{timeDiff >= BonusDeadline 115s?}
    E -->|Yes| F[bonus = MaxBonus 10]
    E -->|No| G["bonus = MaxBonus × (timeDiff / BonusDeadline)"]
    F --> H["score = BaseScore(5) + bonus"]
    G --> H
    H --> I[Insert PlayerAnswer to MongoDB]
    Z --> I
```

## Deployment Architecture (Docker Compose)

```mermaid
graph TB
    subgraph docker["Docker Network: bb-network"]
        subgraph infra["Infrastructure"]
            kafka["kafka:9092\nBitnami Kafka KRaft"]
            pg["bb-postgres:5432\nPostgreSQL 17"]
            redis["bb-redis:6379\nRedis 6.2"]
            mongo1["bb-mongo1:27017"]
            mongo2["bb-mongo2:27017"]
            mongo3["bb-mongo3:27017"]
        end
        
        subgraph monitoring["Monitoring"]
            prometheus["bb-prometheus:9090"]
            grafana["bb-grafana:3000"]
            mongoexp["bb-mongo-exporter:9216"]
        end
        
        subgraph app["Application Services"]
            traefik["bb-traefik:80"]
            auth["bb-auth-service:5000/6000"]
            user["bb-user-service:5001/6001"]
            match["bb-match-service:5002"]
            game["bb-game-service:5003"]
        end
    end

    mongo1 --- mongo2
    mongo2 --- mongo3
    mongoexp --> mongo1
    prometheus --> mongoexp
    grafana --> prometheus
    
    traefik --> auth
    traefik --> user
    traefik --> match
    traefik --> game
    
    user --> pg
    user --> auth
    match --> redis
    match --> kafka
    game --> mongo1
    game --> redis
    game --> kafka
```
