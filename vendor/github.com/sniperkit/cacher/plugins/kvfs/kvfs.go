package kvfscache

/*
import (
	"crypto/tls"
	"golang.org/x/net/context"
	net "net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/conductant/kvfs"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
)

// Cache stores and retrieves data using Badger KV.
type Cache struct {
	store   *store.Store
	root    []string
	handler *handler
	config  *Config
}

type Config struct {
	Engine            string
	Endpoint          *net.URL
	BucketName        string
	BucketPath        string
	MountPath         string
	CertFile          string `flag:"cert, The cert file"`
	KeyFile           string `flag:"key, The key file"`
	CACertFile        string `flag:"ca_cert, The CA cert file"`
	TLS               *tls.Config
	PersistConnection bool
	ConnectionTimeout time.Duration `flag:"timeout,The timeout"`
}

// Sadly libkv doesn't not abstract away the differences in handling the keys and other behaviors
// So we'd have to create something like this to make sure things work across different kvstores.
type handler struct {
	nameFromKey       NameFromKeyFunc
	deleteEmptyParent DeleteEmptyParentFunc
}

func Attach(store *store.Store) *Cache {
	return &Cache{store}
}

func (cache *Cache) View(c context.Context, f func(Context) error) error {
	ctx := c.Context(cache)
	return f(ctx)
}

func (cache *Cache) Update(c context.Context, f func(Context) error) error {
	ctx := cache.Context(c)
	return f(ctx)
}

func (cache *Cache) Context(ctx context.Context) Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return NewContext(ctx, cache.store, cache.root, cache.handler)
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
/*
func Mount(c *Config) *Cache {
	closer, err := kvfs.Mount(c.Endpoint, c.MountPath, &config.Config)
	if err != nil {
		panic(err)
	}
	defer closer.Close()

	blockHere := make(chan interface{})
	<-blockHere
}
*/
/*
func New(c *Config) *Cache {
	endpoint, err := net.Parse(c.Endpoint)
	if err != nil {
		return nil, err
	}
	config := &store.Config{
		Bucket: c.BucketPath,
	}
	if c.CertFile != "" && c.KeyFile != "" && c.CACertFile != "" {
		config.ClientTLS = &store.ClientTLSConfig{
			CertFile:   c.CertFile,
			KeyFile:    c.KeyFile,
			CACertFile: c.CACertFile,
		}
		config.TLS = c.TLS
	}
	config.PersistConnection = c.PersistConnection
	config.ConnectionTimeout = c.ConnectionTimeout
	store, handler, err := getStore(endpoint, config)

	return &Cache{
		store:   store,
		handler: handler,
		config:  c,
	}
}

func getStore(u *net.URL, config *store.Config) (s store.Store, h *handler, err error) {
	hosts := strings.Split(u.Host, ",")
	switch u.Scheme {
	case "zk":
		s, err = libkv.NewStore(store.ZK, hosts, config)
		h = &handler{
			NameFromKey: func(parent string, key string) (name string) {
				// Zk return the name, not the path.  So b in /a/b is just b
				return key
			},
			DeleteEmptyParent: func(store store.Store, key string) error {
				return store.Delete(key)
			},
		}
	case "etcd":
		s, err = libkv.NewStore(store.ETCD, hosts, config)
		h = &handler{
			NameFromKey: func(parent string, key string) (name string) {
				// Etcd returns the absolute path.  So we need to split the path and return the name.
				if filepath.IsAbs(key) {
					key = key[1:]
				}
				return strings.Split(strings.Replace(key, parent+"/", "", 1), "/")[0]

			},
			DeleteEmptyParent: func(store store.Store, key string) error {
				return store.DeleteTree(key)
			},
		}
	case "consul":
		s, err = libkv.NewStore(store.CONSUL, hosts, config)
		h = &handler{
			NameFromKey: func(parent string, key string) (name string) {
				// Consul returns the full path but without the leading '/'.
				return strings.Split(strings.Replace(key, parent+"/", "", 1), "/")[0]
			},
			DeleteEmptyParent: func(store store.Store, key string) error {
				return store.DeleteTree(key)
			},
		}
	default:
		s, err = nil, &ErrNotSupported{u.Scheme}
	}
	return
}
*/

// Close closes the underlying boltdb database.
/*
func (c *Cache) Close() error {
	return c.db
}
*/
