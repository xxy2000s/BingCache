package main

import (
	"bingcache"
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
func main() {
	// var s server
	// http.ListenAndServe("localhost:9999", &s)
	// fmt.Println("cache test")

	// 声明一个名为scores的Group，回调函数设置为如果没命中缓存，则从db中获取数据
	bingcache.NewGroup("scores", 2<<10, bingcache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key ", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:8888"
	// 建立一个新的addr对应的HTTPPool
	peers := bingcache.NewHTTPPool(addr)
	log.Println("bingcache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
