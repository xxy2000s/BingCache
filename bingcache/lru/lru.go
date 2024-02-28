package lru

import "container/list"

// TODO: LRU-K, LFU, ARC
// Cache Definition
type Cache struct {
	maxBytes int64
	nbytes   int64
	ll       *list.List
	cache    map[string]*list.Element
	//存回调函数，可以查看淘汰历史数据
	/*
		keys := make([]string, 0)
		callback := func(key string, value Value) {
			keys = append(keys, key)
		}
		lru := New(int64(10), callback)
	*/
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

// 多态Go版本
// Cache结构体实现了接口Value中的Len()方法, 说明其实现了Value接口
// 声明一个接口实例，它的值可以通过实现了这个接口的结构体赋予（结构体类型强制转换为接口类型）
// 接口实例可以直接调用接口中的方法，这样就实现了多态（多个实现了接口方法的结构体，会通过结构体赋值的不同，调用对应的方法）
func (c *Cache) Len() int {
	return c.ll.Len()
}

// 构造Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// func: find cache, delete cache, add/modify cache
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// 传入的value是一个接口类型，其他任何类型值的结构体或者值传入都可以
// 所有类型的值都实现了Len()
// 传入了之后通过类型断言，就能从接口类型实例转为具体实例
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		//ele.Value 是any类型，等价于interface{}, 是没有方法的空接口
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		//存在回调函数的话，就将该操作加入回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}
