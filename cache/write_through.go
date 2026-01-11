package cache

import (
	"context"
	"log"
	"time"
)

// Write Through
// 开发者只需要写入 cache，cache 自己会更新数据库在读未命中缓存的情况下,开发者需要自己去数据库捞数据，然后更新缓存（此时缓存不需要更新 DB 了）

type WriteThroughCache struct {
	Cache
	StoreFunc func(ctx context.Context, key string, val any) error
}

func (w *WriteThroughCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	err := w.StoreFunc(ctx, key, val)
	if err != nil {
		return err
	}
	return w.Cache.Set(ctx, key, val, expiration)
}

func (w *WriteThroughCache) SetV1(ctx context.Context, key string, val any, expiration time.Duration) error {
	err := w.Cache.Set(ctx, key, val, expiration)
	if err != nil {
		return err
	}
	return w.StoreFunc(ctx, key, val)
}

func (w *WriteThroughCache) SetV2(ctx context.Context, key string, val any, expiration time.Duration) error {
	err := w.StoreFunc(ctx, key, val)
	// 半异步刷缓存
	go func() {
		er := w.Cache.Set(ctx, key, val, expiration)
		if er != nil {
			log.Fatalln(er)
		}
	}()
	return err
}

func (w *WriteThroughCache) SetV3(ctx context.Context, key string, val any, expiration time.Duration) error {
	// 全异步
	go func() {
		err := w.StoreFunc(ctx, key, val)
		if err != nil {
			log.Fatalln(err)
		}
		if err = w.Cache.Set(ctx, key, val, expiration); err != nil {
			log.Fatalln(err)
		}
	}()
	return nil
}

type WriteThroughCacheV1[T any] struct {
	Cache
	StoreFunc func(ctx context.Context, key string, val T) error
}

func (w *WriteThroughCacheV1[T]) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	err := w.StoreFunc(ctx, key, val.(T))
	if err != nil {
		return err
	}
	return w.Cache.Set(ctx, key, val, expiration)
}
