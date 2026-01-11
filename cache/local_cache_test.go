package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildInMapCache_Get(t *testing.T) {
	testCases := []struct {
		name    string
		key     string
		cache   func() *BuildInMapCache
		wantVal any
		wantErr error
	}{
		{
			name: "key not found",
			key:  "not exist key",
			cache: func() *BuildInMapCache {
				return NewBuildInMapCache(10 * time.Second)
			},
			wantErr: errKeyNotFound,
		},
		{
			name: "get val",
			key:  "key1",
			cache: func() *BuildInMapCache {
				res := NewBuildInMapCache(10 * time.Second)
				err := res.Set(context.Background(), "key1", 123, time.Minute)
				require.NoError(t, err)
				return res
			},
			wantVal: 123,
		},
		{
			name: "expired",
			key:  "expired key",
			cache: func() *BuildInMapCache {
				res := NewBuildInMapCache(10 * time.Second)
				err := res.Set(context.Background(), "expired key", 123, time.Second)
				require.NoError(t, err)
				time.Sleep(time.Second * 2)
				return res
			},
			wantErr: errKeyExpired,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := tc.cache().Get(context.Background(), tc.key)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVal, val)
		})
	}
}

func TestNewBuildInMapCache_Loop(t *testing.T) {
	cnt := 0
	c := NewBuildInMapCache(time.Second, BuildInMapCacheOptionWithEvictedCallBack(func(key string, val any) {
		cnt++
	}))
	err := c.Set(context.Background(), "key1", 123, time.Second)
	require.NoError(t, err)
	time.Sleep(time.Second * 3)
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	_, ok := c.data["key1"]
	require.False(t, ok)
	require.Equal(t, 1, cnt)
}

// Write Back
// 在写操作的时候写了缓存直接返回，不会直接更新数据 库，读也是直接读缓存
// 在缓存过期的时候，将缓存写回去数据库
// 优缺点：
// 所有 goroutine 都是读写缓存，不存在一致性的问题（如果是本地缓存依旧会有问题）
// 数据可能丢失：如果在缓存过期刷新到数据库之前，缓存宕机，那么会丢失数据
func TestNewBuildInMapCache_WriteBack(t *testing.T) {
	NewBuildInMapCache(time.Second, BuildInMapCacheOptionWithEvictedCallBack(func(key string, val any) {
		// 刷新到数据库
	}))
}
