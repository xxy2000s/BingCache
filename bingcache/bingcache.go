package bingcache

import (
	"fmt"
	"log"
	"mycache/singleflight"
	"sync"
)

// 定义接口
type Getter interface {
	Get(key string) ([]byte, error)
}

// 声明函数，参数和返回值都与接口中的方法一致
type GetterFunc func(key string) ([]byte, error)

// 接口型函数 --> 调用时既能传入函数作为参数，又能用实现了该接口的结构体作为参数
// 定义一个函数类型 F，并且实现接口 A 的方法，然后在这个方法中调用自己。这是 Go 语言中将其他函数（参数返回值定义与 F 一致）转换为接口 A 的常用技巧。
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// 一个Group是一个命名空间，对应一种类型的缓存，在并发缓存的外层再套上缓存名及数据源的回调函数
type Group struct {
	name      string //缓存名
	getter    Getter //缓存未命中时获取源数据的回调(callback)
	mainCache cache  //实现的并发缓存
	peers     PeerPicker
	loader    *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	// 空处理
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	// 命中缓存
	if v, ok := g.mainCache.Get(key); ok {
		log.Println("[BingCache] hit")
		return v, nil
	}
	// 没命中，从数据源加载
	return g.load(key)
}

// load函数，选择不同的加载数据源方式
func (g *Group) load(key string) (value ByteView, err error) {
	// 外层包上 Do 函数，从而请求实现并发
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[BingCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

// 将实现了 PeerPicker 接口的 HTTPPool 注入Group
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// 从远程节点加载源数据
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

// 从本地加载源数据
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// 将访问的数据加入缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.Add(key, value)
}
