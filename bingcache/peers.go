package bingcache

// 通过一致性哈希中的节点名称key选择对应的任一虚拟哈希节点？
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// 选择到的Peer应该具备的功能(通过group和key获取缓存)
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
