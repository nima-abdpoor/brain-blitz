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
