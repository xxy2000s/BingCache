## Cache
[7天用Go从零实现分布式缓存GeeCache | 极客兔兔](https://geektutu.com/post/geecache.html)

仿groupcache:https://github.com/golang/groupcache

**cache运行流程**

```cpp
                            是
接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
                |  否                         是
                |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
                            |  否
                            |-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶

```

GeeCache支持：
- 单机缓存与基于http的分布式缓存
- LRU淘汰策略
- go锁机制防止缓存击穿
- 一致性hash选择节点，实现负责均衡
- protobuf优化节点间二进制通信

缓存应用：
1. 访问一个网页，网页和引用的 JS/CSS 等静态文件，根据不同的策略，会缓存在浏览器本地或是 CDN 服务器
2. 微博的点赞的数量，缓存在redis集群

缓存的问题：
1. 内存不够了怎么办？
  1. 随机删除数据（淘汰策略：按时间？按频率？）
2. 并发写入冲突？
  1. go的map没有并发，需要加锁
3. 单机性能不够？
  1. 水平扩展（scale horizontally）：多机分布式（节点通信？http?tcp?rpc?)（节点选择？一致性hash实现负载均衡）
  2. 垂直扩展（scale vertically）：增加单节点计算、存储、带宽等性能
