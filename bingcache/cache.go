package bingcache

import (
	"bingcache/lru"
	"sync"
)

// 实例化LRU缓存，并封装get add等功能
type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

// 可以传lru.Value 也可以传实际结构体ByteView
func (c *cache) Add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 延迟初始化，等到lru对象被使用时才开始创建实例
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	// lru.Add(string, Value) 形参是接口, 可以传递实现了该接口的结构体对象
	c.lru.Add(key, value)
}

func (c *cache) Get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if value, ok := c.lru.Get(key); ok {
		return value.(ByteView), ok
	}
	return
}
