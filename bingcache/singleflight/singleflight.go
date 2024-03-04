package singleflight

import "sync"

// call表示正在进行中，或者已结束的请求
type call struct {
	wg  sync.WaitGroup // 用信道的话接受和发送需要一一对应，但请求可能有多个同时wait()
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex // 保护 m
	m  map[string]*call
}

// 对应相同的 key，无论 Do被调用多少次，fn只会被调用一次
// 针对多个同时请求相同 key的情形，可以只等待第一个请求完成（请求的缓冲器）
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	// 如果存在请求便等待第一个请求完成，并直接拿到返回值（全局置入了group的call中）
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()
	c.val, c.err = fn()
	c.wg.Done()
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
	return c.val, c.err

}
