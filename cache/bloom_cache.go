package cache

import "context"

type BloomFilter interface {
	HasKey(ctx context.Context, key string) bool
}
type BloomFilterCache struct {
	ReadThroughCache
}

// NewBloomFilterCache 无侵入式
func NewBloomFilterCache(cache Cache, bf BloomFilter,
	loadFunc func(ctx context.Context, key string) (any, error)) *BloomFilterCache {
	return &BloomFilterCache{
		ReadThroughCache: ReadThroughCache{
			Cache: cache,
			LoadFunc: func(ctx context.Context, key string) (any, error) {
				if !bf.HasKey(ctx, key) {
					return nil, errKeyNotFound
				}
				return loadFunc(ctx, key)
			},
		},
	}
}

type BloomFilterCacheV1 struct {
	ReadThroughCache
	Bf BloomFilter
}

// Get 侵入式
func (r *BloomFilterCacheV1) Get(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound && r.Bf.HasKey(ctx, key) {
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
