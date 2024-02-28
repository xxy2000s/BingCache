package bingcache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGetter(t *testing.T) {
	// 声明一个具体的接口型函数（有具体功能）
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failded")
	}
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	bing := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			// 正在查找
			log.Println("[SlowDB] search key", key)
			// db中存在数据
			if v, ok := db[key]; ok {
				// 没有调用过回调函数
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil
			}
			// db中不存在该数据
			return nil, fmt.Errorf("%s not exists", key)
		}))

	for k, v := range db {
		// 缓存数据和数据库中的数据不一致
		if view, err := bing.Get(k); err != nil || view.String() != v {
			t.Fatalf("failed to get value of %s", k)
		}
		// 多次调用了回调函数
		if _, err := bing.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	// 对不存在的数据进行查询
	if view, err := bing.Get("NotExist"); err == nil {
		fmt.Print("test")
		t.Fatalf("NotExist value doesn't exist, but %s got", view)
	}

}

func TestGetGroup(t *testing.T) {
	groupName := "scores"
	NewGroup(groupName, 2<<10, GetterFunc(func(key string) (bytes []byte, err error) {
		return
	}))
	if group := GetGroup(groupName); group == nil || group.name != groupName {
		t.Fatalf("group %s not exists", groupName)
	}
	if group := GetGroup(groupName + "111"); group != nil {
		t.Fatalf("expect nil, but %s got", group.name)
	}
}
