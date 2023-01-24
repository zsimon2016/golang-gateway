package server

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	util "gitee.com/simon_git_code/goutil"
	// libkvStore "github.com/rpcxio/libkv/store"
	etcd_client "github.com/rpcxio/rpcx-etcd/client"
	etcdv3 "github.com/rpcxio/rpcx-etcd/store/etcdv3"
	"github.com/smallnest/rpcx/client"
)

var (
	etcdAddr   = flag.String("EtcdAddr", "localhost:2379", "Etcd service address")
	basePath   = flag.String("base", "/services", "Prefix path")
	configPath = flag.String("config", "./config.json", "Configuration file path")
)

type Service struct {
	ConfMap  map[string]interface{}
	Services map[string]client.XClient
	Redis    *util.Redis
}

func Init() Service {
	flag.Parse()
	s := Service{}
	services := make(map[string]client.XClient)
	store, err := etcdv3.New([]string{*etcdAddr}, nil)
	// strSliceHeader := *(*reflect.StringHeader)(unsafe.Pointer(&config))
	// byteSlice := *(*[]byte)(unsafe.Pointer(&strSliceHeader))
	// opt := libkvStore.WriteOptions{
	// 	IsDir: false,
	// 	TTL:   1000000 * time.Hour,
	// }
	// store.Put("apiConf", byteSlice, &opt)
	if rf, err := ioutil.ReadFile(*configPath); err == nil {
		json.Unmarshal(rf, &s.ConfMap)
	} else {
		log.Println("Configuration file parsing error:", err)
	}
	if err == nil {
		path := "/services/"
		sl, e := store.List(path)
		if e != nil {
			return s
		}
		for _, v := range sl {
			if path+string(v.Value) == string(v.Key) {
				svr := string(v.Value)
				d, _ := etcd_client.NewEtcdV3Discovery(*basePath, svr, []string{*etcdAddr}, false, nil)
				c := client.NewXClient(svr, client.Failtry, client.RandomSelect, d, client.DefaultOption)
				services[svr] = c
			}
		}
	}
	s.Services = services
	// fmt.Println("=======", s.ConfMap)
	rConf := (s.ConfMap)["redis"].(map[string]interface{})
	rhost := rConf["host"].(string)
	rpwd := rConf["password"].(string)
	pool := util.GetRedisPool(&rhost, &rpwd)
	redis := util.Redis{pool}
	s.Redis = &redis
	return s
}
func (s *Service) ServicePool(svr string) client.XClient {
	if s.Services[svr] == nil {
		d, _ := etcd_client.NewEtcdV3Discovery(*basePath, svr, []string{*etcdAddr}, false, nil)
		c := client.NewXClient(svr, client.Failtry, client.RandomSelect, d, client.DefaultOption)
		s.Services[svr] = c
	}
	return s.Services[svr]
}
