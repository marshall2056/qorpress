package badgercache

import (
	"github.com/dgraph-io/badger"
	// "github.com/sniperkit/httpcache/helpers"
)

const (
	bktDir      = "./shared/data/cache/badger"
	bktValueDir = "values"
	bktName     = "httpcache"
)

// Cache stores and retrieves data using Badger KV.
type Cache struct {
	db *badger.DB
}

type Config struct {
	Dir        string
	ValueDir   string
	SyncWrites bool
}

func Attach(db *badger.DB) *Cache {
	return &Cache{db}
}

func New(config *Config) *Cache {
	badgerConfig := &badger.DefaultOptions{}
	if config == nil {
		badgerConfig.Dir = bktDir
		badgerConfig.ValueDir = bktValueDir
		badgerConfig.SyncWrites = false
	} else {
		badgerConfig.Dir = config.Dir
		badgerConfig.ValueDir = config.ValueDir
		badgerConfig.SyncWrites = config.SyncWrites
	}
	return &Cache{
		db: badger.Open(badgerConfig),
	}
}

// Close closes the underlying boltdb database.
func (c *Cache) Close() error {
	return c.db.Close()
}
