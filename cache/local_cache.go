package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	errKeyNotFound = errors.New("cache: 键不存在")
	errKeyExpired  = errors.New("cache: 键过期")
)

type BuildInMapCacheOption func(cache *BuildInMapCache)

type BuildInMapCache struct {
	data  map[string]*item
	mutex sync.RWMutex
	close chan struct{}
	// 回调与关闭 CDC(change data capture) 在key被更新的时候打印一些数据
	onEvicted func(key string, val any)
	//onEvicted func(ctx context.Context, key string, val any)
	//onEvicts  []func(key string, val any)
}

func NewBuildInMapCache(interval time.Duration, opts ...BuildInMapCacheOption) *BuildInMapCache {
	res := &BuildInMapCache{
		data:  make(map[string]*item, 100),
		close: make(chan struct{}),
		onEvicted: func(key string, val any) {

		},
	}

	for _, opt := range opts {
		opt(res)
	}
	// 轮询的方式处理过期时间
	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case t := <-ticker.C:
				res.mutex.Lock()
				i := 0
				for key, val := range res.data {
					if i > 1000 {
						break
					}
					if val.deadlineBefore(t) {
						res.delete(key)
					}
					i++
				}
				res.mutex.Unlock()
			case <-res.close:
				return
			}
		}
	}()
	return res
}

func BuildInMapCacheOptionWithEvictedCallBack(fn func(key string, val any)) BuildInMapCacheOption {
	return func(cache *BuildInMapCache) {
		cache.onEvicted = fn
	}
}

func (b *BuildInMapCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.set(key, val, expiration)
}

func (b *BuildInMapCache) set(key string, val any, expiration time.Duration) error {
	var dl time.Time
	if expiration > 0 {
		dl = time.Now().Add(expiration)
	}
	b.data[key] = &item{
		val:      val,
		deadline: dl,
	}
	return nil
}

// 这种相当于每个key开一个goroutine盯着，过期了就删除
// 缺点：key多，G多，而且这些G大部分时候是阻塞的，会浪费资源
//func (b *BuildInMapCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
//	b.mutex.Lock()
//	defer b.mutex.Unlock()
//	var dl time.Time
//	if expiration > 0 {
//		dl = time.Now().Add(expiration)
//	}
//	b.data[key] = &item{
//		val:      val,
//		deadline: dl,
//	}
//	// 认为如果 expiration 小于等于0 就是永不过期
//	if expiration > 0 {
//
//		// 用一个结构体 item 就不会出现下面注释的情况
//		// 这种相当于每个key开一个goroutine盯着，过期了就删除
//		// 缺点：key多，G多，而且这些G大部分时候是阻塞的，会浪费资源
//		time.AfterFunc(expiration, func() {
//			b.mutex.Lock()
//			defer b.mutex.Unlock()
//			val, ok := b.data[key]
//			if ok && !val.deadline.IsZero() && val.deadline.Before(time.Now()) {
//				delete(b.data, key)
//			}
//		})
//
//		// 假如说你第十秒设置了 key1=value1 过期时间一分钟
//		// 然后30s 设置了 key1=value2 过期时间一分钟 但是前面的一分钟一到，就会删除掉key1
//		//time.AfterFunc(expiration, func() {
//		//	b.mutex.Lock()
//		//	defer b.mutex.Unlock()
//		//	delete(b.data, key)
//		//})
//	}
//
//	return nil
//}

// redis中也是这种玩法
// get的时候检查是否过期
// 遍历过期key找出过期的删掉，但是也会控制遍历的次数，防止资源的开销
func (b *BuildInMapCache) Get(ctx context.Context, key string) (any, error) {
	b.mutex.RLock()
	res, ok := b.data[key]
	b.mutex.RUnlock()
	if !ok {
		return nil, errKeyNotFound
	}
	now := time.Now()
	// 尝试删数据，肯定要加锁
	// 从读锁到这个锁之间可能会被别人拿了锁，拿了锁可能会存在数据不一致，所以就要double-check
	if res.deadlineBefore(now) {
		b.mutex.Lock()
		defer b.mutex.Unlock()
		// double-check，查要删除的key还是不是原先的key
		res, ok = b.data[key]
		if !ok {
			return nil, errKeyNotFound
		}
		if res.deadlineBefore(now) {
			// 删除是因为如果是G轮询查看过期时间，有可能这个key早过期了，但是还没轮训到，用户这个时候用Get查就会有脏数据
			b.delete(key)
			return nil, errKeyExpired
		}
	}
	return res.val, nil
}

func (b *BuildInMapCache) Delete(ctx context.Context, key string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.delete(key)
	return nil
}

func (b *BuildInMapCache) delete(key string) {
	itm, ok := b.data[key]
	if !ok {
		return
	}
	delete(b.data, key)
	b.onEvicted(key, itm.val)
}

// Close 要是用户调用两次怎么办？
// 可以帮用户做这个判断，也可以不帮，因为用户你就不该调用两次
func (b *BuildInMapCache) Close() error {
	b.close <- struct{}{}
	return nil
}

type item struct {
	val      any
	deadline time.Time
}

func (i *item) deadlineBefore(t time.Time) bool {
	return !i.deadline.IsZero() && i.deadline.Before(t)
}
