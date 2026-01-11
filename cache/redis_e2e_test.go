//go:build e2e

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestRedisCache_e2e_Set(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "43.154.97.245:6379",
	})
	c := NewRedisCache(rdb)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := c.Set(ctx, "key1", "value1", time.Minute)
	require.NoError(t, err)
	res, err := c.Get(ctx, "key1")
	require.NoError(t, err)
	require.Equal(t, "value1", res)
}

func TestRedisCache_e2e_SetV1(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "43.154.97.245:6379",
	})
	testCases := []struct {
		name string
		// 准备数据
		before func()
		// 清理数据
		after      func(t *testing.T)
		key        string
		value      string
		expiration time.Duration
		wantErr    error
	}{
		{
			name:   "set value",
			before: func() {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				res, err := rdb.Get(ctx, "key1").Result()
				require.NoError(t, err)
				require.Equal(t, "value1", res)
				// 不能影响其他测试，所以最后要删除掉
				_, err = rdb.Del(ctx, "key1").Result()
				require.NoError(t, err)
			},
			key:        "key1",
			value:      "value1",
			expiration: time.Minute,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewRedisCache(rdb)
			tc.before()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			err := c.Set(ctx, tc.key, tc.value, tc.expiration)
			require.NoError(t, err)
			tc.after(t)
		})
	}
}
