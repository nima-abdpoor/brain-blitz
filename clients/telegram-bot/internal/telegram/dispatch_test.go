package telegram

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"testing"
	"time"
)

func TestSlowChatDoesNotBlockAnotherChat(t *testing.T) {
	s := &countdownSender{events: make(chan tgbotapi.Chattable, 4)}
	b := New(s, testLogger())
	blocked := make(chan struct{})
	entered := make(chan struct{})
	b.RegisterCommand("start", func(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
		if msg.Chat.ID == 1 {
			close(entered)
			select {
			case <-blocked:
			case <-ctx.Done():
			}
		}
		return tgbotapi.NewMessage(msg.Chat.ID, "hello"), nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	updates := make(chan tgbotapi.Update, 2)
	done := make(chan struct{})
	go func() { defer close(done); b.Run(ctx, updates) }()
	updates <- tgbotapi.Update{Message: commandMessage(1, "start")}
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("first chat did not start")
	}
	updates <- tgbotapi.Update{Message: commandMessage(2, "start")}
	select {
	case c := <-s.events:
		if c.(tgbotapi.MessageConfig).ChatID != 2 {
			t.Fatal("wrong chat replied")
		}
	case <-time.After(time.Second):
		t.Fatal("second chat was blocked")
	}
	close(blocked)
	cancel()
	<-done
}
