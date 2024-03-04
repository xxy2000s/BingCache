package main

import (
	"bingcache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// // 收到Request，并给予回复的句柄
// type server int
//
//	func (h *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//		log.Println(r.URL.Path)
//		w.Write([]byte("Hello World"))
//	}

// 创建缓存 Group
func createGroup() *bingcache.Group {
	return bingcache.NewGroup("scores", 2<<10, bingcache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s doesn't exist", key)
		}))
}

// 启动缓存服务器：创建 HTTPPool，添加节点信息，注册到 bing 中，启动 HTTP 服务
func startCacheServer(addr string, addrs []string, bing *bingcache.Group) {
	// 声明 peers 为 HTTPPool 并初始化，并在 group 空间中注册
	peers := bingcache.NewHTTPPool(addr)
	peers.Set(addrs...)
	bing.RegisterPeers(peers)
	log.Println("bingcache is running at ", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

// 启动一个 API 服务，与用户进行交互
func startAPIServer(apiAddr string, bing *bingcache.Group) {
	http.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		view, err := bing.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(view.ByteSlice())
	}))
	log.Println("fonrtend server is running at ", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	// // var s server
	// // http.ListenAndServe("localhost:9999", &s)
	// // fmt.Println("cache test")

	// // 声明一个名为scores的Group，回调函数设置为如果没命中缓存，则从db中获取数据
	// bingcache.NewGroup("scores", 2<<10, bingcache.GetterFunc(
	// 	func(key string) ([]byte, error) {
	// 		log.Println("[SlowDB] search key ", key)
	// 		if v, ok := db[key]; ok {
	// 			return []byte(v), nil
	// 		}
	// 		return nil, fmt.Errorf("%s not exist", key)
	// 	}))

	// addr := "localhost:8888"
	// // 建立一个新的addr对应的HTTPPool
	// peers := bingcache.NewHTTPPool(addr)
	// log.Println("bingcache is running at", addr)
	// log.Fatal(http.ListenAndServe(addr, peers))
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Bingcache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	bing := createGroup()
	if api {
		go startAPIServer(apiAddr, bing)
	}
	startCacheServer(addrMap[port], []string(addrs), bing)
}
