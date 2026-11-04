# Local Telegram client validation — 2026-11-04

Validated on `telegram-client` after merging `main` at `2506027`.

## Automated results

- Bot unit tests passed with the race detector, followed by `go vet ./...`
  and `go build ./...` in `clients/telegram-bot`.
- Game service and repository tests passed with the race detector; game
  packages also passed `go vet`.
- All four bot integration scenarios passed with the race detector against
  real Docker services. No integration scenario was skipped. The final run
  completed in 50.179 seconds.
- The full-game scenario drove two simulated Telegram chats through the real
  dispatcher, authentication middleware, registration, profile, logout, login,
  category selection, matchmaking, readiness, and ten answer buttons per chat.
  Each chat received ten score updates and final results. Correct answers
  earned points; deliberately incorrect answers earned zero. Both players
  could reconnect after completion.
- Final Docker images built successfully for the bot and game service.

The isolated `brainblitz-telegram-e2e` deployment used `compose.smoke.yaml`
with two-second question windows and the included demo questions. Its
containers were removed after validation; the ordinary local backend retains
its normal game timers.

## Failures found and fixed

1. Traefik discovered containers from both Compose projects and sent some
   requests into the other project's unreachable network. Its Docker provider
   is now restricted to its own Compose project.
2. Immediate READY commands could precede asynchronous question delivery.
   The backend now waits for the questions with a bounded timeout and does not
   schedule completion after a failed start.
3. The bot presented the next question immediately after an accepted answer,
   although the backend rejects answers before the next window opens. It now
   waits until the previous deadline before presenting the next question.

## Live Telegram status

The initial live startup was blocked by a missing token and network timeouts.
After the user configured the token and restored connectivity, the bot started
successfully at 2026-11-04 11:07:50 UTC:

- Telegram authentication succeeded as `@official_brain_blitz_bot`.
- Redis session storage connected successfully.
- The metrics endpoint responded and the bot began long polling for updates.
- The local user and game services remained healthy.

The automated full-game test replaces Telegram's Sender/update-channel boundary.
Live counters subsequently confirmed three `/start` commands, registration and
login actions, a successful profile request, and a MUSIC queue connection.
Some login attempts returned backend HTTP 500; the backend's known behavior
for a nonexistent account is still a generic error rather than invalid
credentials. A complete live match through Telegram's servers still requires
confirmation from players.

Live testing also exposed an SDK response mismatch for button acknowledgements:
`answerCallbackQuery` returns a boolean, whereas the SDK's `Send` helper expects
a Message. An API sender adapter now uses `Request` for these acknowledgements.
Its regression test and the Telegram adapter tests passed with the race
detector, followed by vet and the bot build.

Start the configured bot from the repository root with:

```sh
docker compose --profile telegram up -d bb-telegram-bot
```

Open the bot in Telegram, send `/start`, register two players, select the same
category, and verify questions, answer callbacks, and final results in both
chats. The repeatable automated test commands are in
`clients/telegram-bot/README.md`.

## Follow-up implementation and verification

The real Telegram session subsequently completed. Redis showed both players
FINISHED; MongoDB showed FINISHED, scores of 50 and 45, and nine submitted
answers per player. Unanswered questions expired normally. The log-following
process was stopped after these checks.

The follow-up changes implement server-controlled rounds, cancellation presence,
named standings, personalized outcomes, full-width buttons, local countdowns,
bounded Telegram calls, per-chat dispatch, safe logging, and correct unknown
account errors. The current implementation status is in `live-testing-bugs.md`.

Verification includes:

- Full backend and bot race-test suites, vet, and Linux amd64/arm64 builds.
- Regression checks for concurrent and duplicate submissions, stale timeout
  jobs, repeated READY without resetting a round, idempotent completion,
  cancelled/delayed joins, and cross-category queue membership.
- Bot checks for winner/loser identity independent of leaderboard order, ties,
  public names, countdown edits and cancellation, token redaction, request
  retries, and independent chat dispatch.
- Five real-backend integration scenarios, with no skipped tests: registration
  and profile, wrong password, matchmaking, full Telegram-shaped conversation
  and ten-round game, and unknown-account login. The full suite passed in
  26.586 seconds; the conversation/game scenario passed in 15.07 seconds,
  including onboarding and matchmaking. Both answered rounds advance early.
- A separate full-game scenario with one silent player passed, exercising
  deadline advancement, zero-score final results for the silent player, and
  post-game reconnect. The final silent-player scenario passed in 34.276 seconds with reduced
  deadline scheduler latency.
- The independent backend smoke client passed all ten rounds, final scoring,
  gateway authentication, and reconnect against the same isolated deployment.

The integration tests replace Telegram transport with simulated chats; they do
not prove updated mobile rendering or countdown edits through Telegram's own
servers. Those visual checks remain for the next human playthrough. The local
runtime is updated only after automated checks pass. External Telegram network
availability is still required.

### Final local runtime check

The new game, match, and user images were applied and all backend services
reported healthy. The original MongoDB result remained FINISHED at 50–45.
Redis's temporary per-player status keys had expired normally; the durable
result was retained.

The updated bot image was also applied, but Telegram authentication timed out
again during startup. An independent Docker curl request to api.telegram.org
also failed to connect on port 443 (HTTP 000). The bot remains configured to
restart automatically; live use requires restoring the working network/VPN
route or setting an existing reachable proxy through BOT_HTTPS_PROXY. Startup
errors now redact the token. No live log-following process was restarted.

All changes are local to telegram-client and have not been committed or pushed.

## Follow-up test and new findings — 2026-11-04

- Collected redacted logs for the completed real Telegram test in
  `/private/tmp/brain-blitz-retest-complete.log`; stopped the follower (exit
  130), then confirmed no Compose log-follower processes remained.
- Real game `6aeb32aef14ff359a488fa33` finished with ten answers each and
  scores 44–29. The user confirmed the prior fixes. The only warning in the
  reviewed test logs was an expired callback acknowledgement at startup.
- Implemented profile statistics in game-service, user-service, and bot:
  historical completed games, outcomes, points, rates, and global competition
  rank. New players are unranked; stats outages leave account details usable.
- Added action/category/menu emojis and corrected Help's obsolete statement
  that gameplay was unavailable. Callback values remain unchanged.
- Full bot race tests and vet passed; affected backend tests and vet passed.
  Mongo regression passed against a disposable database for historical totals,
  silent players, draws, tied ranks, new accounts, unfinished-game exclusion,
  and rank updates. Profile propagation/outage and bot rendering tests passed.
- Built and redeployed game-service, user-service, and the Telegram bot.
  Backend services reported healthy. Real registration/login/profile API
  integration passed with zero games and unranked for the fresh account.
- Deployed historical profile for player 13 correctly returned two games,
  two wins, 94 total points, best 50, average 47, and rank 1. Unauthenticated
  gateway requests to the new statistics route returned HTTP 401.
- Telegram initially timed out during deployment. A subsequent one-shot log
  check confirmed successful authentication and update-loop startup at
  13:44:14 UTC. Continuous log monitoring remains stopped.
