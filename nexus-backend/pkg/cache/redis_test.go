package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T) (*RedisClient, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client, err := NewRedisClient(mr.Addr(), "", 0)
	require.NoError(t, err)
	return client, mr
}

func TestNewRedisClient_Success(t *testing.T) {
	mr := miniredis.RunT(t)
	c, err := NewRedisClient(mr.Addr(), "", 0)
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNewRedisClient_ParseURL(t *testing.T) {
	mr := miniredis.RunT(t)
	c, err := NewRedisClient("redis://"+mr.Addr(), "", 0)
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNewRedisClient_PingFail(t *testing.T) {

	_, err := NewRedisClient("127.0.0.1:19998", "", 0)
	assert.Error(t, err)
}

func TestRedisClient_SetAndGet(t *testing.T) {
	c, _ := newTestClient(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "key1", "value1", time.Minute))

	val, err := c.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)
}

func TestRedisClient_Get_Miss(t *testing.T) {
	c, _ := newTestClient(t)
	_, err := c.Get(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestRedisClient_Delete(t *testing.T) {
	c, _ := newTestClient(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "del-key", "val", time.Minute))
	require.NoError(t, c.Delete(ctx, "del-key"))

	_, err := c.Get(ctx, "del-key")
	assert.Error(t, err)
}

func TestRedisClient_Close(t *testing.T) {
	c, _ := newTestClient(t)
	assert.NoError(t, c.Close())
}
