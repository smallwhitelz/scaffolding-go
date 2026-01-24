//go:build e2e

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_e2e_TryLock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "43.154.97.245:6379",
	})

	testCases := []struct {
		name       string
		before     func(t *testing.T)
		after      func(t *testing.T)
		key        string
		expiration time.Duration
		wantErr    error
		wantLock   *Lock
	}{
		{
			// 别人持有了锁
			name: "key exists",
			before: func(t *testing.T) {
				// 模拟别人有锁
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "key1", "value1", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.GetDel(ctx, "key1").Result()
				require.NoError(t, err)
				assert.Equal(t, "value1", res)
			},
			key:        "key1",
			expiration: time.Minute,
			wantErr:    ErrFailedToPreemptLock,
		},
		{
			// 加锁成功
			name:   "locked",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.GetDel(ctx, "key2").Result()
				require.NoError(t, err)
				// 加锁成功意味着你应该设置好了值
				assert.NotEmpty(t, res)
			},
			key:        "key2 ",
			expiration: time.Minute,
			wantLock: &Lock{
				key: "key2",
			},
		},
	}
	client := NewClient(rdb)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			lock, err := client.TryLock(ctx, tc.key, tc.expiration)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantLock.key, lock.key)
			assert.NotEmpty(t, lock.value)
			assert.NotNil(t, lock.client)
			tc.after(t)
		})
	}
}

func TestLock_e2e_Unlock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "43.154.97.245:6379",
	})
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		lock *Lock

		wantErr error
	}{
		{
			name:   "lock not hold",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {

			},
			lock: &Lock{
				key:    "unlock_key1",
				value:  "123",
				client: rdb,
			},
			wantErr: ErrLockNotHold,
		},
		{
			name: "lock hold by others",
			before: func(t *testing.T) {
				// 模拟别人的锁
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "unlock_key2", "value2", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				// 没释放锁，所以键值对不变
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.GetDel(ctx, "unlock_key2").Result()
				require.NoError(t, err)
				assert.Equal(t, "value2", res)
			},
			lock: &Lock{
				key:    "unlock_key2",
				value:  "123",
				client: rdb,
			},
			wantErr: ErrLockNotHold,
		},
		{
			name: "unlocked",
			before: func(t *testing.T) {
				// 模拟自己加的锁
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "unlock_key3", "123", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				// 锁被释放，key不存在
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Exists(ctx, "unlock_key3").Result()
				require.NoError(t, err)
				assert.Equal(t, int64(0), res)
			},
			lock: &Lock{
				key:    "unlock_key3",
				value:  "123",
				client: rdb,
			},
			wantErr: ErrLockNotHold,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()
			err := tc.lock.Unlock(ctx)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}
