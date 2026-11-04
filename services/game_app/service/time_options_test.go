package service

import (
	"BrainBlitz.com/game/pkg/timeoptions"
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

type optionProvider struct{ options []timeoptions.Option }

func (p optionProvider) List(context.Context) ([]timeoptions.Option, error) { return p.options, nil }
func TestServerValidatesTimeOption(t *testing.T) {
	svc := Service{}.WithTimeOptions(optionProvider{[]timeoptions.Option{{ID: "10-seconds", Label: "10 seconds", Seconds: 10, Enabled: true}, {ID: "disabled", Label: "Old", Seconds: 60, Enabled: false}}})
	c, err := svc.resolveTimeOption(context.Background(), "10-seconds")
	require.NoError(t, err)
	require.Equal(t, 10, c.Seconds)
	_, err = svc.resolveTimeOption(context.Background(), "disabled")
	require.Error(t, err)
	_, err = svc.resolveTimeOption(context.Background(), "missing")
	require.Error(t, err)
	legacy, err := svc.resolveTimeOption(context.Background(), "")
	require.NoError(t, err)
	require.Zero(t, legacy.Seconds)
}
