# BrainBlitz Android client roadmap

BrainBlitz for Android uses the shared accounts, matchmaking queues, and game server. Connection setup, account screens, foreground gameplay, and match recovery are implemented. Local emulator validation passed. Staging validation and beta packaging remain open. [Run the Android app](../../clients/android/README.md)

The Gradle app under `clients/android/` uses Kotlin, Jetpack Compose, a ViewModel with `StateFlow`, and OkHttp for HTTP and WebSockets. Networking and session storage sit outside the UI, and explicit state types model matchmaking and gameplay. [Android app module](../../clients/android/app) · [Telegram WebSocket protocol](../../clients/telegram-bot/internal/apiclient/ws/protocol.go) · [Go smoke client](../../tools/smoke/main.go)

## Existing backend contracts

| Capability | Gateway contract | Android behavior |
| --- | --- | --- |
| Account | `POST /user-service/public/api/v1/signup` and `POST /user-service/public/api/v1/login` | Use existing BrainBlitz credentials. Login returns the player ID and tokens. |
| Profile | `GET /user-service/api/v1/profile` | Show the display name and available stats. Handle an unranked player or missing stats. |
| Match settings | `GET /question-service/api/v1/question/time-options`, `GET /game-service/api/v1/game-options`, and the initial socket frame when available | Fetch timer choices and player limits. Read categories from the initial `{categories, numberOfPlayers}` frame for a fresh session. |
| Match and game | WebSocket upgrade at `/game-service/api/v1/process-game` | Send `ADD_TO_WAITING_LIST`, `LEAVE_QUEUE`, `READY`, `ANSWER`, and `SYNC_MATCH`. Handle match, question, answer feedback, round recap, completion, and recovery events. |

Use `Authorization: Bearer <accessToken>` for protected HTTP requests and the WebSocket handshake. Compose already routes the player endpoints above. Verify the selected staging ingress early, especially if it uses the checked-in Kubernetes manifest, which lacks a question-service route. [Compose gateway](../../docker-compose.yml) · [Kubernetes ingress](../../infra/kubernetes/ingress/ingress-route.yaml)

## 1. Prove the connection

**Implemented.**

Create the Gradle project, debug and staging endpoint configuration, and a minimal network client. Use seeded Docker Compose services for local development. The Android emulator can reach the default local gateway through `http://10.0.2.2:8088`. Limit cleartext traffic to local debug builds and use HTTPS and WSS on staging. Add an Android CI job for the debug build, lint, and unit tests. [Local setup and smoke test](../run-on-another-computer.md) · [Android emulator networking](https://developer.android.com/studio/run/emulator-networking-address)

**Done when** an emulator can log in with a disposable account, fetch the profile and match options, and receive the initial WebSocket frame through Traefik.

## 2. Build account and matchmaking screens

**Implemented.**

Add signup, login, profile, and sign-out. Persist the access token in app-private storage encrypted with an Android Keystore key and redact credentials and tokens from logs. For the first milestone, prompt for sign-in when the access token expires. Login returns a refresh token, and the current gateway exposes no client token-exchange endpoint. [User response types](../../services/user_app/service/param.go) · [Android Keystore](https://developer.android.com/privacy-and-security/keystore)

Players select a category, timer, and total player count. Android sends the selection and display name with `ADD_TO_WAITING_LIST`, prevents duplicate taps, and opens the ready screen on `MATCH_CREATED`. A failed join stays visible as an error even if an `ADDED_TO_WAITING_LIST` frame follows. Queue cancellation uses the confirmation described in phase 4. [WebSocket command handler](../../services/game_app/service/websocket_handler.go)

**Done when** two distinct accounts using identical settings enter the same match, and a cancelled player can select new settings and requeue.

## 3. Complete the foreground game

**Implemented.** Two Android clients and a mixed Android–Telegram pair completed seeded 10-question matches with consistent standings.

Implement readiness, question and answer screens, the server-deadline countdown, immediate personal feedback from `ANSWER_ACCEPTED`, the `ROUND_CLOSED` recap, and final standings from `COMPLETED`. Keep match state across rotation, disable duplicate or expired submissions, and let the server determine round transitions and scores. Test protocol decoding with fixtures and cover both a full round of answers and a player who lets the timer expire. [Game event payloads](../../services/game_app/service/param.go) · [Server round progression](../../services/game_app/service/rounds.go)

**Done when** two Android clients finish a seeded match, see consistent standings, and can play again. Repeat with one Android player and one Telegram player using distinct accounts to verify interoperability.

## 4. Make interrupted matches recoverable

**Implemented.** Local tests covered recovery during readiness and play, after connectivity loss, and after completion while Android was backgrounded.

The authenticated game socket accepts `SYNC_MATCH` and returns `MATCH_SNAPSHOT` for the account's latest match. This works without receiving `MATCH_CREATED` or saving a game ID on the device. The game service orders the snapshot with match creation and round updates. The client consumes later frames after applying the snapshot. Finished results are recoverable for 24 hours, or until the player starts another match. [Game socket contract](backend-api-analysis.md#recovering-a-match)

The game service defaults to a 60-second ready window and emits `MATCH_CANCELLED` when it expires. `LEAVE_QUEUE` returns `LEFT_WAITING_LIST` when cancellation wins, or `MATCH_CREATED` when the match already exists. Android checks for a match at sign-in and process startup, and retries on network changes, socket failure, or app resume. It restores the original deadline and uses the server's submission status before enabling an answer. The sign-in-again flow is sufficient for recovery in this phase. Persistent sign-in beyond expiry would require a client token-exchange endpoint and refresh-token lifecycle decisions.

**Done when** a player can recover after backgrounding, changing networks, or restarting the app during a match, and sees the ready screen, current round, or final result as appropriate. Verify that recovery preserves the original deadline and scores an answer at most once.

## 5. Prepare a beta build

- [x] Local emulator testing passed on two Android 10 emulators against the Docker backend. Coverage included expired sessions, queue cancellation, ready timeout, a 15-player lobby, missed answers, draws, match recovery, large text, and timer readability.
- [ ] Validate against staging on an emulator and a physical device, including a TalkBack pass.
- [ ] Produce a signed beta build and document installation, environment selection, and troubleshooting with redacted logs.

**Done when** a tester can install the build and complete or recover a match against staging. If Google Play is the distribution channel, include its current target SDK, Data safety, privacy policy, and in-app and web account-deletion request paths in the release scope. [Google Play target API requirements](https://developer.android.com/google/play/requirements/target-sdk) · [Google Play account deletion requirements](https://support.google.com/googleplay/android-developer/answer/13327111?hl=en)
