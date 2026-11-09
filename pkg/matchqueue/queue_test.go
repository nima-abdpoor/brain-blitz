package matchqueue

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestCancelAndCategoryChangeRejectDelayedJoins(t *testing.T) {
	m := miniredis.RunT(t)
	c := redis.NewClient(&redis.Options{Addr: m.Addr()})
	defer c.Close()
	ctx := context.Background()
	require.NoError(t, Join(ctx, c, "1", "SPORT"))
	require.NoError(t, Enqueue(ctx, c, "waiting_users", "1", "SPORT"))
	require.Equal(t, int64(1), c.ZCard(ctx, "waiting_users:SPORT").Val())
	require.NoError(t, Leave(ctx, c, "waiting_users", "1"))
	require.NoError(t, Enqueue(ctx, c, "waiting_users", "1", "SPORT"))
	require.Equal(t, int64(0), c.ZCard(ctx, "waiting_users:SPORT").Val())
	require.NoError(t, Join(ctx, c, "1", "MUSIC"))
	require.NoError(t, Enqueue(ctx, c, "waiting_users", "1", "MUSIC"))
	require.NoError(t, Enqueue(ctx, c, "waiting_users", "1", "SPORT"))
	require.Equal(t, int64(0), c.ZCard(ctx, "waiting_users:SPORT").Val())
	require.Equal(t, int64(1), c.ZCard(ctx, "waiting_users:MUSIC").Val())
}

func TestTimeSelectionRejectsStaleJoin(t *testing.T) {
	m := miniredis.RunT(t)
	c := redis.NewClient(&redis.Options{Addr: m.Addr()})
	defer c.Close()
	ctx := context.Background()
	fast := Choice{Seconds: 10, OptionID: "10-seconds", Players: 2}
	slow := Choice{Seconds: 60, OptionID: "1-minute", Players: 2}
	require.NoError(t, Join(ctx, c, "1", "SPORT", fast))
	require.NoError(t, Join(ctx, c, "1", "SPORT", slow))
	require.NoError(t, Enqueue(ctx, c, "waiting_users", "1", "SPORT", fast))
	require.Equal(t, int64(0), c.ZCard(ctx, "waiting_users:SPORT").Val())
	require.NoError(t, Enqueue(ctx, c, "waiting_users", "1", "SPORT", slow))
	require.Equal(t, int64(1), c.ZCard(ctx, "waiting_users:SPORT").Val())
	require.True(t, Active(ctx, c, "1", "SPORT", slow))
	require.False(t, Active(ctx, c, "1", "SPORT", fast))
	choice, err := ChoiceFor(ctx, c, "1")
	require.NoError(t, err)
	require.Equal(t, slow, choice)
	require.NoError(t, Leave(ctx, c, "waiting_users", "1"))
	require.False(t, Active(ctx, c, "1", "SPORT"))
}

func TestPlayerCountSelectionRejectsDelayedJoinAndCancellation(t *testing.T) {
	m := miniredis.RunT(t)
	c := redis.NewClient(&redis.Options{Addr: m.Addr()})
	defer c.Close()
	ctx := context.Background()
	old := Choice{Seconds: 10, OptionID: "fast", Players: 3}
	selected := Choice{Seconds: 10, OptionID: "fast", Players: 15}
	require.NoError(t, Join(ctx, c, "1", "TECH", old))
	require.NoError(t, Join(ctx, c, "1", "TECH", selected))
	require.NoError(t, Enqueue(ctx, c, "waiting_users", "1", "TECH", old))
	require.Zero(t, c.ZCard(ctx, "waiting_users:TECH").Val())
	require.False(t, Active(ctx, c, "1", "TECH", old))
	require.True(t, Active(ctx, c, "1", "TECH", selected))
	require.NoError(t, Enqueue(ctx, c, "waiting_users", "1", "TECH", selected))
	got, err := ChoiceFor(ctx, c, "1")
	require.NoError(t, err)
	require.Equal(t, selected, got)
	require.NoError(t, Leave(ctx, c, "waiting_users", "1"))
	require.NoError(t, Enqueue(ctx, c, "waiting_users", "1", "TECH", selected))
	require.Zero(t, c.ZCard(ctx, "waiting_users:TECH").Val())
}
