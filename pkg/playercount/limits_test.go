package playercount

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLimits(t *testing.T) {
	l := (Limits{}).WithDefaults()
	require.NoError(t, l.Validate())
	require.Equal(t, 2, l.Min)
	require.Equal(t, 15, l.Max)
	require.Len(t, l.Options(), 14)
	for n := 2; n <= 15; n++ {
		got, err := l.Resolve(n)
		require.NoError(t, err)
		require.Equal(t, n, got)
	}
	for _, n := range []int{-1, 1, 16} {
		_, err := l.Resolve(n)
		require.Error(t, err)
	}
	l = Limits{Min: 3, Max: 4}
	require.Equal(t, []int{3, 4}, l.Options())
	got, err := l.Resolve(0)
	require.NoError(t, err)
	require.Equal(t, 3, got)
	_, err = l.Resolve(2)
	require.Error(t, err)
	for _, l := range []Limits{{Min: 1, Max: 15}, {Min: 4, Max: 3}, {Min: -1, Max: 15}} {
		require.Error(t, l.Validate())
	}
}
