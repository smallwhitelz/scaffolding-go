package cache

import (
	"context"
	"errors"
	"log"
	"time"

	"golang.org/x/sync/singleflight"
)

// read_through业务代码只需要从 cache 中读取数据，cache 会在缓存不命中的时候去读取数据
// 写数据的时候，业务代码需要自己写 DB 和写 cache

var (
	ErrFailedToRefreshCache = errors.New("刷新缓存失败")
)

// ReadThroughCache 你一定要赋值 LoadFunc 和 Expiration
// Expiration 是你的过期时间
type ReadThroughCache struct {
	Cache
	LoadFunc   func(ctx context.Context, key string) (any, error)
	Expiration time.Duration
	sg         singleflight.Group
}

func (r *ReadThroughCache) Get(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		val, err = r.LoadFunc(ctx, key)
		if err == nil {
			er := r.Cache.Set(ctx, key, val, r.Expiration)
			if er != nil {
				return val, ErrFailedToRefreshCache
			}
		}
	}
	return val, err
}

// GetV1 全异步
func (r *ReadThroughCache) GetV1(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		go func() {
			val, err = r.LoadFunc(ctx, key)
			if err == nil {
				er := r.Cache.Set(ctx, key, val, r.Expiration)
				if er != nil {
					log.Fatalln(er)
				}
			}
		}()
	}
	return val, err
}

// GetV2 半异步
func (r *ReadThroughCache) GetV2(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		val, err = r.LoadFunc(ctx, key)
		if err == nil {
			go func() {
				er := r.Cache.Set(ctx, key, val, r.Expiration)
				if er != nil {
					log.Fatalln(er)
				}
			}()
		}
	}
	return val, err
}

// GetV3
func (r *ReadThroughCache) GetV3(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		val, err, _ = r.sg.Do(key, func() (interface{}, error) {
			v, er := r.LoadFunc(ctx, key)
			if er == nil {
				er = r.Cache.Set(ctx, key, val, r.Expiration)
				if er != nil {
					log.Fatalln(er)
				}
			}
			return v, er
		})
	}
	return val, err
}

type ReadThroughCacheV1[T any] struct {
	Cache
	LoadFunc   func(ctx context.Context, key string) (T, error)
	Expiration time.Duration
	sg         singleflight.Group
}

func (r *ReadThroughCacheV1[T]) Get(ctx context.Context, key string) (T, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		val, err = r.LoadFunc(ctx, key)
		if err == nil {
			er := r.Cache.Set(ctx, key, val, r.Expiration)
			if er != nil {
				return val.(T), ErrFailedToRefreshCache
			}
		}
	}
	return val.(T), err
}
