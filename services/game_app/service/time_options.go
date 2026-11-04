package service

import (
	"BrainBlitz.com/game/pkg/matchqueue"
	"BrainBlitz.com/game/pkg/timeoptions"
	"context"
	"fmt"
)

func (svc Service) WithTimeOptions(provider timeoptions.Provider) Service {
	svc.timeOptions = provider
	return svc
}
func (svc Service) resolveTimeOption(ctx context.Context, id string) (matchqueue.Choice, error) {
	// Older clients remain in a separate legacy pool with the configured timer.
	if id == "" {
		return matchqueue.Choice{}, nil
	}
	if svc.timeOptions == nil {
		return matchqueue.Choice{}, fmt.Errorf("question time options unavailable")
	}
	options, err := svc.timeOptions.List(ctx)
	if err != nil {
		return matchqueue.Choice{}, fmt.Errorf("question time options unavailable")
	}
	for _, option := range options {
		if option.ID == id && option.Enabled && option.Validate() == nil {
			return matchqueue.Choice{Seconds: option.Seconds, OptionID: option.ID}, nil
		}
	}
	return matchqueue.Choice{}, fmt.Errorf("question time option is no longer available; choose again")
}
