package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash函数
type Hash func(data []byte) uint32

type Map struct {
	hash     Hash           //Hash函数
	replicas int            //虚拟节点倍数
	keys     []int          //hash环
	hashMap  map[int]string //虚拟节点（用字符串经过hash之后得到的值）和真实节点（字符串）映射关系
}

// 一致性Hash算法的主数据结构
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	// 默认Hash函数
	if fn == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			//算出hash
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	// 二分搜索找到m.keys数组中第一个大于hash值的节点索引
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	// 因为二分搜索的索引可能等于len(m.keys) 所以要取余
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
