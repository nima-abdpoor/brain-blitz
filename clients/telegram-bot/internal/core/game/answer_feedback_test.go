package game

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/apiclient/ws"
)

func feedbackEvent(gameID, questionID string, correct bool) ws.Event {
	data, _ := json.Marshal(ws.AnswerAcceptedMetaData{
		GameID:         gameID,
		AnswerFeedback: &ws.AnswerFeedback{QuestionID: questionID, IsCorrect: correct},
		LeaderBoard:    ws.LeaderboardMetaData{GameID: gameID, PlayerPoint: []ws.PlayerPoint{{PlayerID: "u1", Point: 0}}},
	})
	return ws.Event{Success: true, Event: ws.EventAnswerAccepted, MetaData: data}
}

func roundQuestionEvent(round int, questionID string) ws.Event {
	data, _ := json.Marshal(ws.QuestionsPublishedMetaData{GameID: "g1", RoundNumber: round, TotalQuestions: 2, Questions: []ws.Question{{ID: questionID, Choices: []string{"a", "b"}, TTL: time.Now().Add(time.Hour)}}})
	return ws.Event{Success: true, Event: ws.EventQuestionsPublished, MetaData: data}
}

func TestAnswerFeedbackNotifiesImmediatelyWithoutAdvancingRound(t *testing.T) {
	for _, correct := range []bool{true, false} {
		t.Run(fmt.Sprint(correct), func(t *testing.T) {
			mgr, notifier, _ := newTestManager(t)
			onEvent := mgr.Attached(testChatID, "g1", newFakeConn())
			t.Cleanup(func() { mgr.Disconnected(testChatID) })
			onEvent(roundQuestionEvent(1, "q1"))
			mustRecvQuestion(t, notifier.questions)
			if err := mgr.SubmitAnswer(testChatID, "q1", 0); err != nil {
				t.Fatal(err)
			}
			select {
			case <-notifier.answerResults:
				t.Fatal("reported correctness before server acceptance")
			default:
			}
			onEvent(feedbackEvent("g1", "q1", correct))
			select {
			case got := <-notifier.answerResults:
				if got != correct {
					t.Fatalf("correctness = %v, want %v", got, correct)
				}
			default:
				t.Fatal("no immediate answer feedback")
			}
			board := <-notifier.leaderboards
			if len(board.Scores) != 1 || board.Scores[0].Points != 0 {
				t.Fatalf("leaderboard changed: %+v", board)
			}
			select {
			case <-notifier.questions:
				t.Fatal("feedback advanced the round before the server opened it")
			default:
			}
			onEvent(feedbackEvent("g1", "q1", correct))
			select {
			case <-notifier.answerResults:
				t.Fatal("duplicate response repeated the verdict")
			default:
			}
			onEvent(roundQuestionEvent(2, "q2"))
			if got := mustRecvQuestion(t, notifier.questions); got.question.ID != "q2" {
				t.Fatalf("next question = %s", got.question.ID)
			}
		})
	}
}

func TestAnswerFeedbackIgnoresWrongGameQuestionAndMalformedData(t *testing.T) {
	mgr, notifier, _ := newTestManager(t)
	onEvent := mgr.Attached(testChatID, "g1", newFakeConn())
	t.Cleanup(func() { mgr.Disconnected(testChatID) })
	onEvent(roundQuestionEvent(1, "q1"))
	mustRecvQuestion(t, notifier.questions)
	if err := mgr.SubmitAnswer(testChatID, "q1", 0); err != nil {
		t.Fatal(err)
	}
	for _, event := range []ws.Event{
		feedbackEvent("other-game", "q1", true),
		feedbackEvent("g1", "old-question", true),
		{Event: ws.EventAnswerAccepted, MetaData: []byte(`{"answerFeedback":`)},
	} {
		onEvent(event)
		select {
		case <-notifier.answerResults:
			t.Fatalf("unexpected feedback for %s", event.MetaData)
		default:
		}
	}
	// Invalid responses must leave the real submission pending.
	onEvent(feedbackEvent("g1", "q1", false))
	select {
	case correct := <-notifier.answerResults:
		if correct {
			t.Fatal("wrong verdict")
		}
	default:
		t.Fatal("invalid responses consumed the pending answer")
	}
	<-notifier.leaderboards
	onEvent(roundQuestionEvent(2, "q2"))
	mustRecvQuestion(t, notifier.questions)
	if err := mgr.SubmitAnswer(testChatID, "q2", 1); err != nil {
		t.Fatal(err)
	}
	onEvent(feedbackEvent("g1", "q1", true))
	onEvent(feedbackEvent("g1", "q2", false))
	if len(notifier.answerResults) != 1 || <-notifier.answerResults {
		t.Fatal("stale feedback consumed or overwrote the next question's result")
	}
}
