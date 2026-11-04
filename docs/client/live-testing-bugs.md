# Telegram live-testing bugs and tasks

Observed on 2026-11-04 on `telegram-client`, using two real Telegram players.
The live session completed; MongoDB and Redis confirmed FINISHED for both
players and scores of 50–45. Each player submitted nine answers; unanswered
questions expired. The live log follower was stopped before implementation.

The user subsequently authorized implementing these findings. The table below
supersedes the original observations and pending notes retained further down.

| Item | Implementation |
|---|---|
| 1. Round advancement | Server opens one round at a time; all accepted answers or the deadline advance exactly once; completion is idempotent. |
| 2. Player names | Telegram public names accompany the authenticated queue join and appear in score payloads and standings. |
| 3. Answer labels | One full-width button per row. |
| 4. Connection handling | Separate bounded poll/callback/message timeouts; per-chat workers; bounded retries for explicit Telegram rejections. Ambiguous message deliveries are not retried automatically. |
| 5. Token logging | HTTP errors redact the token while preserving error causes; login/signup error logs omit credentials. |
| 6. False errors | First-answer MongoDB misses and missing initial user state are treated as expected cases. |
| 7. Queue cleanup | Shared Redis presence and atomic membership changes reject delayed cancelled joins and cross-category ghosts; old socket events cannot mutate a replacement bot session. |
| 8. Login errors | Unknown accounts return invalid credentials without backend-outage retries. |
| 9. Progress | Safe round, answer, delivery, and completion logs; answer-recorded progress text. |
| 10. Countdown | Bot calculates remaining time locally and edits Telegram every five seconds; stops on answer, deadline, completion, or disconnect. No backend polling. |
| 11. Outcomes | Per-player winner/loser/draw text and Play again button, using authenticated player identity. |

Cancellation removes a player who is still queued. A disconnect after a match
has already formed follows readiness/round timeouts; it does not resume an
in-progress game on reconnect. Active-game recovery after a process restart
remains outside this implementation, as documented in the client README.

Automated verification is recorded in `local-e2e-validation.md`. Actual mobile
rendering and live countdown edits still need a follow-up chat check after the
new runtime is started.

## 1. Advance the round when all players answer — high priority

**Confirmed behavior:** Both answers to question 1 were scored at 11:21:20 and
11:21:26 UTC, but question 2 did not open until 11:22:56 UTC. Both players then
answered question 2 successfully. The configured answer window is 122 seconds;
answering early currently does not shorten that window.

**Expected behavior, clarified by the user:** As soon as every player has
submitted an accepted answer to the current question, send the next question
to all players immediately. Wait for the deadline only when a player has not
answered.

**Tasks:**

- Make the backend authoritative for current round, start time, and deadline.
- Count distinct players' accepted answers; duplicate taps must not advance
  the round twice. Incorrect but accepted answers count as submissions.
- Advance exactly once when all players answer, or when the deadline expires.
- Open the next answer window and notify all players together. Changing only
  the bot would cause the backend to reject early answers to future questions.
- Finish immediately after the last round closes. Update/cancel scheduled
  completion tasks so the existing full-game timer cannot complete a game twice.
- Decide how disconnected players affect round completion.
- Test concurrent final submissions, duplicate answers, one silent player,
  deadline/submission races, consistent next questions for both players, and
  early completion after the last question.

References: `clients/telegram-bot/internal/core/game/game.go`,
`services/game_app/service/answer_processor.go`,
`services/game_app/service/websocket_handler.go`,
`services/game_app/repository/game.go`.

## 2. Show player names in leaderboards — high priority

**Confirmed behavior:** The leaderboard displays “Player 13” and “Player 14”.
The backend's player-score payload contains `playerId`, points, and answers,
without a display name. The bot explicitly formats this as `Player <id>`.

**Expected behavior:** Show each player's display name in live and final
leaderboards. Internal IDs should remain identifiers in the protocol.

**Tasks:**

- Supply public display names with the match roster or leaderboard entries.
- Carry the name through the WebSocket DTO, bot domain score, and renderer.
- Use a safe fallback when a name is missing, and distinguish duplicate names.
- Avoid displaying account email addresses as an accidental fallback.
- Test names for both opponents in live updates and final results.

References: `services/game_app/service/param.go`,
`clients/telegram-bot/internal/apiclient/ws/protocol.go`,
`clients/telegram-bot/internal/core/game/game.go`,
`clients/telegram-bot/internal/telegram/notifier.go`.

## 3. Answer labels are truncated — confirmed UI bug

The user's screenshot shows “Wrestling” truncated when all four choices share
one inline-keyboard row.

**Task:** Give each answer its own full-width row and verify on mobile with
long choices. Preserve callback question IDs and choice indices.

**Current status:** Implemented and tested with layout-independent integration
selection. The live session has ended and the user authorized applying the
findings. New question messages use the wider layout; historical messages
retain their old keyboards unless explicitly edited.

Reference: `clients/telegram-bot/internal/telegram/keyboards/keyboards.go`.

## 4. Telegram connection failures and stale callbacks

**Confirmed logs:** `getUpdates` and `answerCallbackQuery` experienced
`unexpected EOF`; several callbacks were rejected as too old or expired.

**Tasks:** Bound Telegram request duration, ensure a slow acknowledgement
cannot indefinitely block all chats, and handle transient delivery failures.
Inspect timing before deciding whether stale callbacks originate from network
delay, queued updates, synchronous handling, or a combination. The logs alone
do not establish a single root cause.

References: `clients/telegram-bot/cmd/bot/main.go`,
`clients/telegram-bot/internal/telegram/bot.go`,
`clients/telegram-bot/internal/telegram/notifier.go`.

## 5. Telegram errors expose the bot token in logs — high priority

**Confirmed behavior:** SDK HTTP errors include the Telegram request URL,
which embeds the bot token. No token value is reproduced in this document.

**Tasks:** Redact credentials in SDK and application error logging, including
startup, polling, callback acknowledgements, and message delivery failures.
Add a regression check using a fake token and simulated network errors.

## 6. Normal answer insertion is logged as an error

**Confirmed behavior:** The duplicate-answer lookup logs
`mongo: no documents in result` at ERROR level for a first answer. The code
then continues with insertion; this message does not itself mean the answer
failed. Both real players' question 1 and question 2 answers were scored.

**Task:** Treat `mongo.ErrNoDocuments` as the expected first-submission case;
log actual database errors distinctly and handle them appropriately.

Reference: `services/game_app/repository/game.go`, `SavePlayerAnswer`.

## 7. Cancelled/disconnected queue entries can create unusable matches

**Confirmed logs:** An earlier MUSIC match included a player without a live
socket and logged `MATCH_CREATED` delivery failure (`id ... not found`).
The existing cancel implementation closes the client socket without removing
the backend waiting-list entry. Restarts also lose in-flight bot state.

**Tasks:** Implement backend queue removal/expiry tied to connection lifecycle,
prevent stale or cross-category queue entries from forming unusable matches,
and define recovery behavior after bot restarts. Reproduce cancel/rejoin and
disconnect races explicitly before choosing the final implementation.

## 8. Login failures need useful error classification

Live logs included repeated backend HTTP 500 responses for login. Separately,
source inspection confirms that a missing account is currently classified as
an internal server error. The precise cause of each live failed login was not
verified.

**Tasks:** Return invalid-credentials responses for unknown accounts and wrong
passwords, avoid retrying those failures as outages, and keep actual database
or auth-service failures distinguishable. Do not log passwords.

## 9. Improve progress and delivery observability

**Tasks:** Explain when a player is waiting for an opponent's answer or a round
deadline. Add safe structured logs for round advancement, question delivery
success/failure, accepted answers, and completion. Do not include credentials
or unnecessary account data. This makes “question missing” distinguishable
from “question not open yet” without relying on screenshots.

## 10. Show a locally calculated question countdown

**Requested behavior:** Show the remaining time ticking down in the question
message without polling the BrainBlitz backend.

**Tasks:**

- Use the absolute question deadline already supplied in the WebSocket event.
  Calculate remaining time locally in the bot; no periodic backend request is
  needed.
- Update the visible message through Telegram's message-edit API, preserving
  answer buttons. Normal Telegram messages cannot run their own client-side
  countdown; visible ticking still requires Telegram API calls.
- Consider edits every 3–5 seconds rather than every second, and handle
  Telegram rate limits and transient network errors.
- Retain the outgoing message ID, avoid overlapping edits, and stop updates
  when the player answers, the round ends, the game completes, or the chat's
  connection is abandoned.
- Reconcile the deadline with authoritative round advancement when all players
  answer early (item 1). Timer ticks must not delay incoming answer handling.

## 11. Personalize the final result for winners and losers

**Confirmed live result:** The bot displayed only a generic final leaderboard:
Player 13 scored 50 points and Player 14 scored 45. Player 13 won by 5 points,
but neither chat received an explicit, personalized outcome.

**Expected behavior:** Tell each player whether they won, lost, or drew, then
show the final standings using display names (item 2).

**Tasks:**

- Determine the outcome from the authoritative final leaderboard and the
  current chat's authenticated BrainBlitz user ID, not the Telegram chat ID.
- Send the winner a message such as “🏆 You won! Your score: 50. Opponent: 45.”
- Send the losing player a message such as “Good game! You scored 45; your
  opponent won with 50. Try again!”
- Handle equal top scores explicitly as a draw/shared win according to the
  game's rules, rather than arbitrarily choosing the first leaderboard entry.
- Include final standings and a Play Again button for both players. Avoid
  sending either a winning or losing claim if final standings are incomplete.
- Test winner and loser chat mapping, ties, reversed leaderboard order, and
  duplicate completion events so the result is delivered once per player.

Reference: `clients/telegram-bot/internal/telegram/notifier.go`,
`NotifyCompleted`, and the final leaderboard/session identity mapping.

## Follow-up live test — 2026-11-04

The user confirmed that the previously reported problems are addressed after
playing the redeployed bot. Game `6aeb32aef14ff359a488fa33` completed at
13:29:11 UTC. MongoDB confirms FINISHED, ten answers per player, and scores
of 44 (player 13) and 29 (player 14). Logs show immediate advancement after
both accepted answers and delivery of all ten rounds to both chats.

One stale callback acknowledgement warning occurred immediately after bot
startup at 13:26:31 UTC, before this match. No further WARN/ERROR entries were
found in the reviewed bot/game/match/user logs through completion. This does
not independently verify every visual detail; the user confirmed the fixes.

## 12. Include game statistics and ranking in the profile — implemented

**User finding:** Profile shows account information but omits games played,
scores, and the player's rank.

**Confirmed cause:** The user-service profile response contains only ID,
username, display name, role, and account timestamps. The bot's profile DTO
and renderer also contain only account details. This requires backend data
and bot changes, rather than a text-only adjustment.

**Proposed profile information:**

- Completed games played, wins, losses, draws, and win rate.
- Lifetime points, average points per completed game, and best game score.
- Global rank by lifetime points, with the ranked player count and an explicit
  unranked state before the first completed game. Equal scores should share
  the same rank; document the ranking policy.
- Answers submitted, correct answers, and answer accuracy.

**Implementation tasks:**

- Expose authenticated statistics derived from persisted completed games and
  their answers, including historical games and players with zero answers.
  Exclude unfinished games and avoid counting repeated completion events.
- Keep the game service responsible for game statistics; connect them to the
  profile API and carry the new fields through the bot client/domain/renderer.
- Define ties consistently with final game outcomes; show zero values and
  unranked clearly for new players. Distinguish unavailable stats from zeros.
- Verify known historical totals, wins/losses/draws, tied ranks, no-game users,
  authentication, and updates after a newly completed game.

Implemented after the user ended the test and authorized addressing the new
findings. The profile API includes a `statistics` object, or null when the game
service is unavailable. The bot renders all proposed totals and rates, with
competition ranking (1, 1, 3), and an unranked state for new players.

## 13. Consistent action emojis — implemented for next release

**User request:** Add appropriate emojis for different bot actions in the
next release.

**Tasks:**

- Use a consistent emoji and text label for each action across commands,
  inline buttons, prompts, and notifications: 🎮 Play, 👤 Profile,
  🏆 Rankings, 📊 Statistics, ❓ Help, 🔐 Log in, 📝 Register,
  🚪 Log out, ⬅️ Back, and ❌ Cancel.
- Apply matching cues to matchmaking (🔎), question timers (⏳), accepted
  answers (✅), errors (⚠️), and game outcomes (🏆 win, 🤝 draw).
  An accepted answer must not imply that it is correct before scoring.
- Keep text labels alongside emojis, preserve callback values, and verify
  button readability and emoji rendering on Telegram mobile clients.

Implemented for existing actions in menus, category selection, registration,
login, logout, queue waiting/cancellation, feedback, and results. Button
callback data is unchanged. Help now describes available gameplay and profile
statistics instead of incorrectly saying gameplay is a future feature.

## Previously fixed items

- Early READY before asynchronous question delivery: backend now waits with
  a bounded timeout and does not schedule immediate completion after failure.
- Boolean callback acknowledgement decoded as a Message: the deployed sender
  adapter uses the SDK's Request method for callback acknowledgements.
- Cross-project Traefik discovery: gateways are scoped to their Compose project.

The original ten-question live match finished with verified stored results.
The follow-up behavior is covered by automated tests described in
`local-e2e-validation.md`; updated mobile rendering needs a new playthrough.


## Follow-up test closed and logs collected

The live follower was stopped (exit 130); a process check confirmed no remaining
Compose log followers. Redacted bot/game/match/user logs from 13:26 UTC onward
were saved locally to `/private/tmp/brain-blitz-retest-complete.log`. The only
warning was the expired startup callback already described above. No additional
runtime defect was reproduced in that completed test. No server process was
stopped by ending log collection.

New findings 12–13 and the obsolete Help message were addressed after collection.
Verification and deployment results are in `local-e2e-validation.md`.
