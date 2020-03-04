package libkvcache

/*
import (
	"log"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
)

// logFatal is defined so log.Fatal calls can be overridden for testing
var logFatal = log.Fatal

// LibKV -
type Cache struct {
	store *store.Store
}

// Login -
func (c *Cache) Login() error {
	return nil
}

// Logout -
func (c *Cache) Logout() {
}

func (c *Cache) New(path string) *Cache {
	_, _ = libkv.NewStore(store.ZK, hosts, config)
	return &Cache{}
}

// Read -
func (c *Cache) Read(path string) ([]byte, error) {
	data, err := c.store.Get(path)
	if err != nil {
		return nil, err
	}
	return data.Value, nil
}

func Register(engine string) *Cache {
	switch engine {
	case "consul":
		consul.Register()
	case "etcd":
		etcd.Register()
	case "zookeeper":
		zookeeper.Register()
	}
	return &Cache{}
}
*/
