//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	apihttp "github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/apiclient/http"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/apiclient/ws"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/auth"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/game"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/matchmaking"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/profile"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/session"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/telegram"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/telegram/commands"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/telegram/conversation"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/telegram/keyboards"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/telegram/middleware"
)

// Only the Telegram transport is replaced: updates pass through the real
// dispatcher, conversations, notifiers, HTTP clients, and WebSocket game.
// No Telegram credentials or messages to real chats are needed.
type chatRecorder struct {
	messages chan tgbotapi.MessageConfig
}

func (r *chatRecorder) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	if msg, ok := c.(tgbotapi.MessageConfig); ok {
		r.messages <- msg
	}
	return tgbotapi.Message{}, nil
}

type chatHarness struct {
	t        *testing.T
	ctx      context.Context
	updates  chan tgbotapi.Update
	recorder *chatRecorder
}

func (h *chatHarness) send(chatID int64, text string) {
	h.t.Helper()
	msg := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, Text: text}
	if strings.HasPrefix(text, "/") {
		msg.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(text)}}
	}
	h.update(tgbotapi.Update{Message: msg})
}

func (h *chatHarness) callback(chatID int64, data string) {
	h.t.Helper()
	h.update(tgbotapi.Update{CallbackQuery: &tgbotapi.CallbackQuery{
		ID: fmt.Sprintf("test-%d-%d", chatID, time.Now().UnixNano()), Data: data,
		From:    &tgbotapi.User{FirstName: fmt.Sprintf("Tester %d", chatID)},
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}},
	}})
}

func (h *chatHarness) update(update tgbotapi.Update) {
	h.t.Helper()
	select {
	case h.updates <- update:
	case <-h.ctx.Done():
		h.t.Fatal("bot update loop did not accept the update")
	}
}

func (h *chatHarness) expect(chatID int64, text string) tgbotapi.MessageConfig {
	h.t.Helper()
	timer := time.NewTimer(12 * time.Second)
	defer timer.Stop()
	for {
		select {
		case msg := <-h.recorder.messages:
			if msg.ChatID != chatID {
				h.t.Fatalf("unexpected reply to chat %d while waiting for chat %d", msg.ChatID, chatID)
			}
			if !strings.Contains(msg.Text, text) {
				h.t.Fatalf("reply = %q, want it to contain %q", msg.Text, text)
			}
			return msg
		case <-timer.C:
			h.t.Fatalf("timed out waiting for %q in chat %d", text, chatID)
		case <-h.ctx.Done():
			h.t.Fatal("bot integration test timed out")
		}
	}
}

func TestIntegration_TelegramConversationAndFullGame(t *testing.T) {
	if os.Getenv("BOT_INTEGRATION_FULL_GAME") != "1" {
		t.Skip("set BOT_INTEGRATION_FULL_GAME=1 against an isolated backend with compose.smoke.yaml and seed-demo")
	}
	requireBackend(t, userServiceURL())
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()

	log := discardLogger()
	sessions := session.NewMemoryStore()
	userClient := apihttp.NewUserClient(apihttp.NewClient(userServiceURL(), 10*time.Second, log))
	authSvc := auth.NewService(userClient, sessions)
	recorder := &chatRecorder{messages: make(chan tgbotapi.MessageConfig, 256)}
	dialer := func(ctx context.Context, token string) (matchmaking.Connection, error) {
		return ws.Dial(ctx, gameServiceWSURL(), token)
	}
	matchMgr := matchmaking.NewManager(dialer, telegram.NewMatchmakingNotifier(recorder, log), sessions, nil)
	gameMgr := game.NewManager(telegram.NewGameNotifier(recorder, log, sessions), matchMgr, log)
	matchMgr.SetGameHandler(gameMgr)
	matchMgr.SetTimeOptionsAPI(apihttp.NewTimeOptionsClient(apihttp.NewClient(strings.TrimSuffix(userServiceURL(), "/user-service")+"/question-service", 10*time.Second, log)))
	handlers := commands.NewHandlers(authSvc, profile.NewService(userClient, sessions), matchMgr, gameMgr, conversation.NewMemoryStore())
	bot := telegram.New(recorder, log)
	bot.RegisterCommand("start", handlers.Start)
	bot.RegisterCommand("help", handlers.Help)
	bot.RegisterCommand("profile", middleware.RequireAuth(authSvc, handlers.Profile))
	bot.RegisterCommand("play", middleware.RequireAuth(authSvc, handlers.Play))
	bot.RegisterCallback(keyboards.CallbackRegister, handlers.RegisterCallback)
	bot.RegisterCallback(keyboards.CallbackLogin, handlers.LoginCallback)
	bot.RegisterCallback(keyboards.CallbackLogout, middleware.RequireAuthCallback(authSvc, handlers.LogoutCallback))
	bot.RegisterCallback(keyboards.CallbackPlay, middleware.RequireAuthCallback(authSvc, handlers.PlayCallback))
	bot.RegisterCallback(keyboards.CallbackCancel, middleware.RequireAuthCallback(authSvc, handlers.CancelCallback))
	for _, category := range matchmaking.Categories {
		bot.RegisterCallback(keyboards.CallbackCategoryPrefix+category, middleware.RequireAuthCallback(authSvc, handlers.CategoryCallback))
	}
	bot.RegisterCallbackPrefix(commands.CallbackTimePrefix, func(ctx context.Context, cb *tgbotapi.CallbackQuery, _ string) (tgbotapi.Chattable, error) {
		return middleware.RequireAuthCallback(authSvc, handlers.TimeOptionCallback)(ctx, cb)
	})
	bot.RegisterCallbackPrefix(keyboards.CallbackAnswerPrefix, handlers.AnswerCallback)
	bot.SetTextHandler(handlers.HandleText)
	updates := make(chan tgbotapi.Update, 64)
	done := make(chan struct{})
	go func() {
		defer close(done)
		bot.Run(ctx, updates)
	}()
	defer func() {
		cancel()
		<-done
	}()

	// Cancelling while the Kafka join is in flight must not leave a ghost
	// opponent for the following two-player TECH game.
	if _, err := authSvc.Register(ctx, 2999, uniqueEmail("cancel"), integrationPassword); err != nil {
		t.Fatal(err)
	}
	if err := matchMgr.JoinQueue(ctx, 2999, "TECH"); err != nil {
		t.Fatal(err)
	}
	if err := matchMgr.Cancel(2999); err != nil {
		t.Fatal(err)
	}
	const chatA, chatB = int64(2001), int64(2002)
	defer matchMgr.Detach(chatA)
	defer matchMgr.Detach(chatB)
	h := &chatHarness{t: t, ctx: ctx, updates: updates, recorder: recorder}
	for _, chatID := range []int64{chatA, chatB} {
		email := uniqueEmail(fmt.Sprintf("chat-%d", chatID))
		h.send(chatID, "/profile")
		h.expect(chatID, "You need to be logged in")
		h.send(chatID, "/start")
		h.expect(chatID, "Welcome to BrainBlitz")
		h.callback(chatID, keyboards.CallbackRegister)
		h.expect(chatID, "email address")
		h.send(chatID, email)
		h.expect(chatID, "send your password")
		h.send(chatID, integrationPassword)
		h.expect(chatID, "Registered and logged in")
		h.send(chatID, "/profile")
		h.expect(chatID, email)
		h.callback(chatID, keyboards.CallbackLogout)
		h.expect(chatID, "logged out")
		h.callback(chatID, keyboards.CallbackLogin)
		h.expect(chatID, "email address")
		h.send(chatID, email)
		h.expect(chatID, "send your password")
		h.send(chatID, integrationPassword)
		h.expect(chatID, "You're logged in")
		h.send(chatID, "/play")
		h.expect(chatID, "Choose a category")
	}
	t.Log("PASS: Telegram registration, protected profile, logout, login, and play menu")
	for _, chatID := range []int64{chatA, chatB} {
		h.callback(chatID, keyboards.CallbackCategoryPrefix+"TECH")
		h.expect(chatID, "How much time")
		optionID := os.Getenv("BOT_INTEGRATION_TIME_OPTION")
		if optionID == "" {
			optionID = "10-seconds"
		}
		h.callback(chatID, commands.CallbackTimePrefix+"TECH|"+optionID)
		h.expect(chatID, "Waiting for opponent")
	}

	gameplayStarted := time.Now()
	recaps := map[int64]int{}
	questions := map[int64]int{}
	leaderboards := map[int64]int{}
	completed := map[int64]bool{}
	for len(completed) < 2 {
		select {
		case msg := <-recorder.messages:
			chatID := msg.ChatID
			switch {
			case strings.Contains(msg.Text, "Opponent found"):
				t.Logf("chat %d matched", chatID)
			case strings.HasPrefix(msg.Text, "❓ Question"):
				questions[chatID]++
				var number, total, seconds int
				if _, err := fmt.Sscanf(msg.Text, "❓ Question %d/%d (%ds to answer):", &number, &total, &seconds); err != nil {
					t.Fatalf("invalid question time: %q", msg.Text)
				}
				expected := 10
				if os.Getenv("BOT_INTEGRATION_TIME_OPTION") == "1-minute" {
					expected = 60
				}
				if seconds > expected || seconds < expected-2 {
					t.Fatalf("question timer=%d want approximately %d", seconds, expected)
				}
				keyboard, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
				var buttons []tgbotapi.InlineKeyboardButton
				for _, row := range keyboard.InlineKeyboard {
					buttons = append(buttons, row...)
				}
				if !ok || len(buttons) < 2 {
					t.Fatalf("question has no usable answer buttons: %#v", msg.ReplyMarkup)
				}
				// Demo questions have their correct answer at index 1.
				choice := 0
				if chatID == chatA {
					choice = 1
				}
				if chatID != chatB || os.Getenv("BOT_INTEGRATION_SILENT_PLAYER") != "1" {
					h.callback(chatID, *buttons[choice].CallbackData)
				}
			case strings.HasPrefix(msg.Text, "📝 Question"):
				recaps[chatID]++
				if recaps[chatID] != questions[chatID] || !strings.Contains(msg.Text, "Your answer:") || !strings.Contains(msg.Text, "Correct answer:") {
					t.Fatalf("invalid recap: %q", msg.Text)
				}
				if chatID == chatB && os.Getenv("BOT_INTEGRATION_SILENT_PLAYER") == "1" && !strings.Contains(msg.Text, "No answer submitted") {
					t.Fatalf("silent player recap: %q", msg.Text)
				}
			case strings.HasPrefix(msg.Text, "📊 Leaderboard"):
				leaderboards[chatID]++
			case strings.Contains(msg.Text, "Answer submitted"):
			case strings.HasPrefix(msg.Text, "🏆 Game over! Final results"):
				for _, playerChat := range []int64{chatA, chatB} {
					sess, err := sessions.Get(ctx, playerChat)
					if err != nil {
						t.Fatal(err)
					}
					prefix := sess.DisplayName + ": "
					_, line, found := strings.Cut(msg.Text, prefix)
					var points int
					if !found {
						t.Fatalf("final results missing player %s: %q", sess.UserID, msg.Text)
					}
					if _, err := fmt.Sscanf(line, "%d pts", &points); err != nil {
						t.Fatalf("invalid final score: %q", msg.Text)
					}
					if (playerChat == chatA && points <= 0) || (playerChat == chatB && points != 0) {
						t.Fatalf("incorrect final score for chat %d: %d", playerChat, points)
					}
				}
				if chatID == chatA && !strings.Contains(msg.Text, "You won!") {
					t.Fatalf("winner not personalized: %q", msg.Text)
				}
				if chatID == chatB && !strings.Contains(msg.Text, "Good game!") {
					t.Fatalf("loser not personalized: %q", msg.Text)
				}
				completed[chatID] = true
			default:
				t.Fatalf("unexpected gameplay reply to chat %d: %q", chatID, msg.Text)
			}
		case <-ctx.Done():
			t.Fatalf("full game timed out: questions=%v leaderboards=%v completed=%v", questions, leaderboards, completed)
		}
	}
	if os.Getenv("BOT_INTEGRATION_SILENT_PLAYER") != "1" && time.Since(gameplayStarted) > 25*time.Second {
		t.Error("both-answer game failed to finish early")
	}
	for _, chatID := range []int64{chatA, chatB} {
		wantUpdates := 10
		if chatID == chatB && os.Getenv("BOT_INTEGRATION_SILENT_PLAYER") == "1" {
			wantUpdates = 0
		}
		if recaps[chatID] != 10 {
			t.Errorf("chat %d received %d recaps, want 10", chatID, recaps[chatID])
		}
		if questions[chatID] != 10 || leaderboards[chatID] != wantUpdates {
			t.Errorf("chat %d received %d questions and %d score updates; want 10 each", chatID, questions[chatID], leaderboards[chatID])
		}
		if gameMgr.StateFor(chatID) != game.StateIdle || matchMgr.StateFor(chatID) != matchmaking.Idle {
			t.Errorf("chat %d did not reset after completion", chatID)
		}
		// The same player must be able to open a new authenticated socket.
		sess, err := sessions.Get(ctx, chatID)
		if err != nil {
			t.Fatal(err)
		}
		conn, err := ws.Dial(ctx, gameServiceWSURL(), sess.AccessToken)
		if err != nil {
			t.Fatalf("reconnect after completion: %v", err)
		}
		_ = conn.Close()
	}
	if !t.Failed() {
		t.Log("PASS: two players, ten answer buttons each, scoring, final results, and reconnect")
	}
}
