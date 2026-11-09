// Package playercount defines the configurable, inclusive game-size range.
package playercount

import "fmt"

type Limits struct {
	Min int `koanf:"min" json:"minPlayers"`
	Max int `koanf:"max" json:"maxPlayers"`
}

func (l Limits) WithDefaults() Limits {
	if l.Min == 0 {
		l.Min = 2
	}
	if l.Max == 0 {
		l.Max = 15
	}
	return l
}

func (l Limits) Validate() error {
	if l.Min < 2 || l.Max < l.Min {
		return fmt.Errorf("player limits require 2 <= min <= max (got %d..%d)", l.Min, l.Max)
	}
	return nil
}

// Resolve preserves clients that omit players, using the configured minimum.
func (l Limits) Resolve(players int) (int, error) {
	if players == 0 {
		players = l.Min
	}
	if players < l.Min || players > l.Max {
		return 0, fmt.Errorf("choose between %d and %d players", l.Min, l.Max)
	}
	return players, nil
}

func (l Limits) Options() []int {
	options := make([]int, 0, l.Max-l.Min+1)
	for n := l.Min; n <= l.Max; n++ {
		options = append(options, n)
	}
	return options
}
