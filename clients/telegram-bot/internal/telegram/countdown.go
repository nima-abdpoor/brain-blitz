package telegram

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type countdown struct{ cancel context.CancelFunc }

func (n *GameNotifier) stopCountdown(chatID int64) {
	n.mu.Lock()
	if c, ok := n.countdowns[chatID]; ok {
		c.cancel()
		delete(n.countdowns, chatID)
	}
	n.mu.Unlock()
}

func (n *GameNotifier) startCountdown(chatID int64, messageID int, deadline time.Time, content string, number, total int, keyboard tgbotapi.InlineKeyboardMarkup) {
	if messageID == 0 {
		return
	}
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	n.mu.Lock()
	if n.countdowns == nil {
		n.countdowns = make(map[int64]countdown)
	}
	n.countdowns[chatID] = countdown{cancel}
	n.mu.Unlock()
	go func() {
		defer cancel()
		interval := n.countdownInterval
		if interval <= 0 {
			interval = 5 * time.Second
		}
		tick := time.NewTicker(interval)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				if ctx.Err() != nil {
					return
				}
				seconds := int(time.Until(deadline).Seconds())
				if seconds <= 0 {
					return
				}
				text := fmt.Sprintf("❓ Question %d/%d (⏳ %ds remaining):\n\n%s", number, total, seconds, content)
				edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
				edit.ReplyMarkup = &keyboard
				if _, err := n.api.Send(edit); err != nil {
					n.logger.Warn("countdown edit failed", "chat_id", chatID, "error", err)
				}
			}
		}
	}()
}
