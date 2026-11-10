package telegram

import (
	"context"
	"fmt"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/session"
	"log/slog"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/game"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/telegram/keyboards"
)

// MatchmakingNotifier turns internal/core/matchmaking's asynchronous
// outcomes (match found, failed join, unexpected disconnect) into Telegram
// messages. It implements matchmaking.Notifier structurally — this package
// deliberately does not import internal/core/matchmaking, since Notifier's
// method signatures need nothing from that package (see
// docs/client/client-architecture.md §1: core must not know about
// Telegram, so the dependency only points the other way, wired together in
// cmd/bot/main.go).
//
// These methods fire from a matchmaking.Manager read-loop goroutine, not
// from Bot's own update-driven dispatch, so unlike Bot.send this type talks
// to the Sender directly rather than through the update-handling path.
type MatchmakingNotifier struct {
	api    Sender
	logger *slog.Logger
}

// NewMatchmakingNotifier builds a MatchmakingNotifier. Both dependencies are
// required, injected by the caller.
func NewMatchmakingNotifier(api Sender, logger *slog.Logger) *MatchmakingNotifier {
	return &MatchmakingNotifier{api: api, logger: logger}
}

// NotifyMatchCreated tells chatID an opponent was found. internal/core/game
// takes over immediately after this (see matchmaking.GameHandler), sending
// READY on the chat's behalf — the next message the user sees is either the
// first question (GameNotifier.NotifyQuestion) or, if the opponent never
// readies up, GameNotifier.NotifyReadyTimeout.
func (n *MatchmakingNotifier) NotifyMatchCreated(chatID int64, gameID string) {
	n.send(chatID, tgbotapi.NewMessage(chatID, "🎮 Players found! Getting ready…"))
}

// NotifyQueueFailed tells chatID its join attempt was rejected (e.g. the
// documented ERROR-then-false-success quirk for an invalid category).
func (n *MatchmakingNotifier) NotifyQueueFailed(chatID int64, reason string) {
	reply := tgbotapi.NewMessage(chatID, "⚠️ Couldn't join the queue, try /play again.")
	reply.ReplyMarkup = keyboards.MainMenu()
	n.send(chatID, reply)
}

// NotifyDisconnected tells chatID its connection dropped while waiting.
func (n *MatchmakingNotifier) NotifyDisconnected(chatID int64) {
	reply := tgbotapi.NewMessage(chatID, "⚠️ Lost connection while waiting for players. Send /play to try again.")
	reply.ReplyMarkup = keyboards.MainMenu()
	n.send(chatID, reply)
}

func (n *MatchmakingNotifier) send(chatID int64, msg tgbotapi.MessageConfig) {
	if _, err := n.api.Send(msg); err != nil {
		n.logger.Error("failed to send matchmaking notification", "chat_id", chatID, "error", err)
	}
}

// GameNotifier turns internal/core/game's gameplay events (questions,
// leaderboard updates, timeouts, final results) into Telegram messages. It
// implements game.Notifier — unlike MatchmakingNotifier, this package does
// import internal/core/game here, since Notifier's methods carry that
// package's domain types (Question, Leaderboard); that's still a
// delivery-layer-depends-on-core dependency, the direction
// docs/client/client-architecture.md §1 expects.
//
// Like MatchmakingNotifier, these methods fire from a
// matchmaking-Connection-owned read-loop goroutine, not from Bot's own
// update-driven dispatch.
type GameNotifier struct {
	mu                sync.Mutex
	countdowns        map[int64]countdown
	countdownInterval time.Duration
	sessions          session.Store
	api               Sender
	logger            *slog.Logger
}

// NewGameNotifier builds a GameNotifier. Both dependencies are required.
func NewGameNotifier(api Sender, logger *slog.Logger, stores ...session.Store) *GameNotifier {
	n := &GameNotifier{api: api, logger: logger, countdowns: make(map[int64]countdown), countdownInterval: 5 * time.Second}
	if len(stores) > 0 {
		n.sessions = stores[0]
	}
	return n
}

// NotifyQuestion presents question with a countdown to its deadline and one
// button per choice.
func (n *GameNotifier) NotifyQuestion(chatID int64, q game.Question, questionNumber, totalQuestions int) {
	seconds := int(time.Until(q.Deadline).Round(time.Second) / time.Second)
	if seconds < 0 {
		seconds = 0
	}
	text := fmt.Sprintf("❓ Question %d/%d (%ds to answer):\n\n%s", questionNumber, totalQuestions, seconds, q.Content)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboards.Answer(q.ID, q.Choices)
	n.stopCountdown(chatID)
	sent, err := n.api.Send(msg)
	if err != nil {
		n.logger.Error("question delivery failed", "chat_id", chatID, "error", err)
		return
	}
	n.logger.Info("question delivered", "chat_id", chatID, "question", questionNumber, "deadline", q.Deadline)
	n.startCountdown(chatID, sent.MessageID, q.Deadline, q.Content, questionNumber, totalQuestions, keyboards.Answer(q.ID, q.Choices))
}

// NotifyLeaderboard reports an updated leaderboard after an accepted
// answer, with the game still in progress.
func (n *GameNotifier) NotifyLeaderboard(chatID int64, board game.Leaderboard) {
	n.stopCountdown(chatID)
	n.send(chatID, tgbotapi.NewMessage(chatID, formatLeaderboard("📊 Leaderboard", board)+"\n📩 Answer recorded. The next round starts when everyone answers or time expires."))
}

func (n *GameNotifier) NotifyAnswerAccepted(chatID int64, isCorrect bool, board game.Leaderboard) {
	n.stopCountdown(chatID)
	result := "❌ Incorrect."
	if isCorrect {
		result = "✅ Correct!"
	}
	n.send(chatID, tgbotapi.NewMessage(chatID, formatLeaderboard("📊 Leaderboard", board)+"\n"+result+"\nThe next round starts when everyone answers or time expires."))
}

// NotifyReadyTimeout tells chatID its opponent never became ready.
func (n *GameNotifier) NotifyReadyTimeout(chatID int64) {
	n.stopCountdown(chatID)
	reply := tgbotapi.NewMessage(chatID, "⏳ Not all players became ready. Send /play to look for a new match.")
	reply.ReplyMarkup = keyboards.MainMenu()
	n.send(chatID, reply)
}

// NotifyCompleted reports the final leaderboard; the game is over.
func (n *GameNotifier) NotifyCompleted(chatID int64, board game.Leaderboard) {
	n.stopCountdown(chatID)
	reply := tgbotapi.NewMessage(chatID, formatLeaderboard("🏆 Game over! Final results", board)+n.outcome(chatID, board))
	reply.ReplyMarkup = keyboards.Results()
	n.send(chatID, reply)
}

// NotifyDisconnected tells chatID its connection dropped mid-game.
func (n *GameNotifier) NotifyDisconnected(chatID int64) {
	n.stopCountdown(chatID)
	reply := tgbotapi.NewMessage(chatID, "⚠️ Lost connection during the game. Send /play to start a new match.")
	reply.ReplyMarkup = keyboards.MainMenu()
	n.send(chatID, reply)
}

// formatLeaderboard uses public names, with a fallback for older backends.
func formatLeaderboard(title string, board game.Leaderboard) string {
	var b strings.Builder
	b.WriteString(title + ":\n")
	counts := map[string]int{}
	for _, score := range board.Scores {
		name := score.DisplayName
		if name == "" {
			name = "Player " + score.PlayerID
		}
		counts[name]++
	}
	for i, s := range board.Scores {
		name := s.DisplayName
		if name == "" {
			name = "Player " + s.PlayerID
		}
		if counts[name] > 1 {
			name += fmt.Sprintf(" (%d)", i+1)
		}
		fmt.Fprintf(&b, "%s: %d pts\n", name, s.Points)
	}
	return b.String()
}

func (n *GameNotifier) send(chatID int64, msg tgbotapi.MessageConfig) {
	if _, err := n.api.Send(msg); err != nil {
		n.logger.Error("failed to send game notification", "chat_id", chatID, "error", err)
	}
}

func (n *GameNotifier) outcome(chatID int64, board game.Leaderboard) string {
	if n.sessions == nil || len(board.Scores) < 2 {
		return ""
	}
	sess, err := n.sessions.Get(context.Background(), chatID)
	if err != nil {
		return ""
	}
	top := board.Scores[0].Points
	own := -1
	count := 0
	for _, s := range board.Scores {
		if s.Points > top {
			top = s.Points
		}
		if s.PlayerID == sess.UserID {
			own = s.Points
		}
	}
	if own < 0 {
		return ""
	}
	for _, s := range board.Scores {
		if s.Points == top {
			count++
		}
	}
	if own == top && count > 1 {
		return fmt.Sprintf("\n🤝 It’s a draw! You scored %d points.", own)
	}
	if own == top {
		return fmt.Sprintf("\n🏆 You won! You scored %d points.", own)
	}
	return fmt.Sprintf("\n👏 Good game! You lost this match. You scored %d points. The winner scored %d. Try again!", own, top)
}
