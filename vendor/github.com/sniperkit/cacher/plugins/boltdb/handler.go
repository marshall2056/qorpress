package boltdbcache

import (
	"errors"
	"path/filepath"
	"sync"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

/*
	Refs:
	- https://github.com/emotionaldots/arbitrage/tree/master/cmd/arbitrage-db
	- https://github.com/toorop/tmail/blob/master/core/scope.go
	- https://github.com/ssut/pocketnpm/blob/renewal/db/models.go
	- https://github.com/SermoDigital/bolt/blob/master/encoding/encoder.go
	- https://github.com/SermoDigital/bolt/blob/master/encoding/util.go
	- https://github.com/SermoDigital/bolt/blob/master/encoding/decoder.go
	- https://github.com/SermoDigital/bolt/tree/master/structures
	- https://github.com/boltdb/bolt/issues/678
	- https://github.com/peter-edge/bolttype-go/blob/master/bolttype.go
	- https://github.com/recoilme/boltapi/blob/master/boltapi.go#L102

	TUI:
	- https://github.com/br0xen/boltbrowser

	WebUI:
	- github.com/evnix/boltdbweb
*/

const (
	bktName string = "httpcache"
	bktDir  string = "./shared/data/cache/.boltdb"
)

var defaultCacheDir string = filepath.Join(bktDir, bktName)

type Cache struct {
	sync.RWMutex
	db          *bolt.DB
	storagePath string
	bucketName  string
	debug       bool
}

type Config struct {
	BucketName  string
	StoragePath string
	Debug       bool
}

func New(config *Config) (*Cache, error) {

	if config == nil {
		log.Println("boltdbcache.New(): config is nil")
		config.StoragePath = defaultCacheDir
	}

	if config.StoragePath == "" {
		return nil, errors.New("boltdbcache.New(): Storage path is not defined.")
	}

	if config.BucketName == "" {
		config.BucketName = bktName
	}

	cache := &Cache{}
	cache.debug = config.Debug

	var err error
	cache.db, err = bolt.Open(config.StoragePath, 0600, nil)
	if err != nil {
		if config.Debug {
			log.WithFields(log.Fields{
				"config": config,
				"cache":  cache,
			}).Fatalf("boltdbcache.New(): Open error: %v", err)
		}
		return nil, err
	}

	init := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(config.BucketName))
		if config.Debug {
			log.WithFields(log.Fields{
				"config": config,
				"cache":  cache,
			}).Fatalf("boltdbcache.New(): CreateBucketIfNotExists error: %v", err)
		}
		return err
	}

	if err := cache.db.Update(init); err != nil {
		if config.Debug {
			log.WithFields(log.Fields{
				"config": config,
				"cache":  cache,
			}).Fatalf("boltdbcache.New(): Update error: %v", err)
		}
		if err := cache.db.Close(); err != nil {
			if config.Debug {
				log.WithFields(log.Fields{
					"config": config,
					"cache":  cache,
				}).Fatalf("boltdbcache.New(): Close error: %v", err)
			}
		}
		return nil, err
	}

	return cache, nil

}

// Mount returns a new Cache using the provided (and opened) bolt database.
func Mount(db *bolt.DB) *Cache {
	return &Cache{db: db}
}

// Close closes the underlying boltdb database.
func (c *Cache) Close() error {
	return c.db.Close()
}

// Get retrieves the response corresponding to the given key if present.
func (c *Cache) Get(key string) (resp []byte, ok bool) {
	// c.RLock()
	// defer c.RUnlock()

	get := func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(c.bucketName))
		if bkt == nil {
			return errors.New("bucket is nil")
		}
		resp = bkt.Get([]byte(key))
		return nil
	}
	if err := c.db.View(get); err != nil {
		if c.debug {
			log.WithFields(log.Fields{
				"bucketName": c.bucketName,
				"key":        key,
				"get":        get != nil,
			}).Fatalf("boltdbcache.Get() View ERROR: ", err)
		}
		return resp, false
	}
	if c.debug {
		log.WithFields(log.Fields{
			"bucketName": c.bucketName,
			"key":        key,
			"get":        get != nil,
			"ok":         resp != nil,
		}).Info("boltdbcache.Get() OK")
	}
	return resp, resp != nil
}

// Set stores a response to the cache at the given key.
func (c *Cache) Set(key string, resp []byte) {
	// c.RLock()
	// defer c.RUnlock()
	// strconv.FormatUint(u.ID, 10)

	set := func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(c.bucketName))
		if bkt == nil {
			return errors.New("bucket is nil")
		}
		return bkt.Put([]byte(key), resp)
	}
	if err := c.db.Update(set); err != nil {
		if c.debug {
			log.WithFields(log.Fields{
				"bucketName": c.bucketName,
				"key":        key,
				"set":        set != nil,
			}).Fatalf("boltdbcache.Set() Update ERROR: ", err)
		}
	}
	if c.debug {
		log.WithFields(log.Fields{
			"bucketName": c.bucketName,
			"key":        key,
			"ok":         set != nil,
		}).Info("boltdbcache.Set() OK")
	}
}

// Delete removes the response with the given key from the cache.
func (c *Cache) Delete(key string) {
	// c.RLock()
	// defer c.RUnlock()

	del := func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(c.bucketName))
		if bkt == nil {
			return errors.New("bucket is nil")
		}
		return bkt.Delete([]byte(key))
	}
	if err := c.db.Update(del); err != nil {
		if c.debug {
			log.WithFields(log.Fields{
				"bucketName": c.bucketName,
				"key":        key,
				"ok":         del != nil,
			}).Fatalf("boltdbcache.Delete() Update ERROR: ", err)
		}
	}
	if c.debug {
		log.WithFields(log.Fields{
			"context":    "Entry",
			"bucketName": c.bucketName,
			"key":        key,
			"ok":         del != nil,
		}).Info("boltdbcache.Set() OK")
	}
}

// Ping connects to the database. Returns nil if successful.
func (c *Cache) Ping() error {
	return c.db.View(func(tx *bolt.Tx) error { return nil })
}
