package telegram

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

type botAPI interface {
	Send(tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

// APISender handles methods whose Telegram response is a boolean rather than
// a Message. The SDK's Send method only decodes Message responses.
type APISender struct{ api botAPI }

func NewAPISender(api botAPI) *APISender { return &APISender{api: api} }

func (s *APISender) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	if _, ok := c.(tgbotapi.CallbackConfig); ok {
		_, err := s.api.Request(c)
		return tgbotapi.Message{}, err
	}
	for attempt := 0; ; attempt++ {
		msg, err := s.api.Send(c)
		if err == nil || attempt >= 2 {
			return msg, err
		}
		var apiErr tgbotapi.Error
		// Only retry explicit Telegram rejection, not ambiguous network
		// failures: a message might already have been delivered in that case.
		if !errors.As(err, &apiErr) || (apiErr.Code != 429 && apiErr.Code < 500) {
			return msg, err
		}
		delay := time.Duration(attempt+1) * time.Second
		if apiErr.RetryAfter > 0 {
			delay = time.Duration(apiErr.RetryAfter) * time.Second
		}
		if delay > 5*time.Second {
			return msg, err
		}
		time.Sleep(delay)
	}
}
