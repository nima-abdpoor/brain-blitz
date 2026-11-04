package telegram

import (
	"encoding/json"
	"errors"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type rejectingAPI struct {
	booleanAPI
	calls int
	fail  error
}

func (a *rejectingAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	a.calls++
	if a.calls == 1 {
		return tgbotapi.Message{}, a.fail
	}
	return tgbotapi.Message{MessageID: 7}, nil
}

func TestAPISenderRetriesExplicitRejectionWithoutRetryingAmbiguousDelivery(t *testing.T) {
	api := &rejectingAPI{fail: tgbotapi.Error{Code: 500, Message: "temporarily unavailable"}}
	if _, err := NewAPISender(api).Send(tgbotapi.NewMessage(1, "hello")); err != nil || api.calls != 2 {
		t.Fatalf("explicit rejection calls=%d err=%v", api.calls, err)
	}
	api = &rejectingAPI{fail: errors.New("connection lost after send")}
	if _, err := NewAPISender(api).Send(tgbotapi.NewMessage(1, "hello")); err == nil || api.calls != 1 {
		t.Fatalf("ambiguous delivery retried: calls=%d err=%v", api.calls, err)
	}
}

type booleanAPI struct{ requests, messages int }

func (a *booleanAPI) Request(tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	a.requests++
	return &tgbotapi.APIResponse{Ok: true, Result: json.RawMessage(`true`)}, nil
}

func (a *booleanAPI) Send(tgbotapi.Chattable) (tgbotapi.Message, error) {
	a.messages++
	return tgbotapi.Message{MessageID: 7}, nil
}

func TestAPISenderAcknowledgesBooleanCallbackResponse(t *testing.T) {
	api := &booleanAPI{}
	sender := NewAPISender(api)
	if _, err := sender.Send(tgbotapi.NewCallback("callback-1", "")); err != nil {
		t.Fatal(err)
	}
	if api.requests != 1 || api.messages != 0 {
		t.Fatalf("callback used requests=%d messages=%d", api.requests, api.messages)
	}
	msg, err := sender.Send(tgbotapi.NewMessage(1, "hello"))
	if err != nil || msg.MessageID != 7 || api.messages != 1 {
		t.Fatalf("message response = %+v, error = %v", msg, err)
	}
}
