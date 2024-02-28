package bingcache

// 封装缓存值，用于支持任何类型的数据
type ByteView struct {
	b []byte
}

// 实现Value接口
func (v ByteView) Len() int {
	return len(v.b)
}

// b只读，通过ByteSlice来获得拷贝值，防止缓冲值被外部程序更改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
