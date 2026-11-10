package telegram

import (
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/game"
)

type countdownSender struct{ events chan tgbotapi.Chattable }

func (s *countdownSender) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	s.events <- c
	return tgbotapi.Message{MessageID: 7}, nil
}

func TestCountdownEditsQuestionAndStopsAfterAnswer(t *testing.T) {
	s := &countdownSender{events: make(chan tgbotapi.Chattable, 100)}
	n := NewGameNotifier(s, testLogger())
	n.countdownInterval = 10 * time.Millisecond
	n.NotifyQuestion(42, game.Question{ID: "q1", Content: "Question", Choices: []string{"a", "b"}, Deadline: time.Now().Add(3 * time.Second)}, 1, 10)
	<-s.events
	select {
	case event := <-s.events:
		edit, ok := event.(tgbotapi.EditMessageTextConfig)
		if !ok || edit.MessageID != 7 || edit.ReplyMarkup == nil {
			t.Fatalf("invalid edit: %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("no countdown edit")
	}
	n.NotifyAnswerAccepted(42, true, game.Leaderboard{})
	time.Sleep(30 * time.Millisecond)
	for len(s.events) > 0 {
		<-s.events
	}
	select {
	case event := <-s.events:
		t.Fatalf("countdown continued after answer: %#v", event)
	case <-time.After(30 * time.Millisecond):
	}
}
