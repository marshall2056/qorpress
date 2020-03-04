//
package boltgzipcache

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	bolt "github.com/zhuharev/boltutils"
)

const (
	bktName                     string = "httpcache"
	bktDir                      string = "./shared/data/cache/.boltgzip"
	defaultStorageFileExtension string = ".bbolt"
)

var (
	defaultStorageDir  string = filepath.Join(bktDir, bktName)
	defaultStorageFile string = fmt.Sprintf("%s/%s%s", defaultStorageDir, bktName, defaultStorageFileExtension)
)

/*
	Refs:
	- https://github.com/br0xen/boltbrowser
*/

// Cache is an implementation of httpcache.Cache that uses a bolt database.
type Cache struct {
	// sync.Mutex
	sync.RWMutex
	db          *bolt.DB
	storagePath string
	bucketName  string
	debug       bool
}

type Config struct {
	BucketName     string
	StoragePath    string
	Compressor     string
	ReadOnly       bool
	StrictMode     bool
	NoSync         bool
	NoFreelistSync bool
	NoGrowSync     bool
	MaxBatchSize   bool
	MaxBatchDelay  bool
	AllocSize      bool
	Debug          bool
}

// New returns a new Cache that uses a bolt database at the given path.
func New(config *Config) (*Cache, error) {

	if config.Debug {
		log.WithFields(log.Fields{
			"config": config,
		}).Warnf("boltgzipcache.New()")
	}

	if config == nil {
		config.StoragePath = defaultStorageFile
		config.BucketName = bktName
		config.Compressor = "gzip"
	}

	if config.StoragePath == "" {
		config.StoragePath = defaultStorageFile
	}

	if config.BucketName == "" {
		config.BucketName = bktName
	}

	if config.Debug {
		log.WithFields(log.Fields{
			"config": config,
		}).Warnf("boltgzipcache.New() ---> post-processed")
	}

	cache := &Cache{}
	cache.storagePath = config.StoragePath
	cache.bucketName = config.BucketName
	cache.debug = config.Debug

	var err error
	switch strings.ToLower(config.Compressor) {
	case "lz4":
		cache.db, err = bolt.New(bolt.OpenPath(config.StoragePath), bolt.Compression(bolt.Lz4Compressor))
	case "gzip":
		fallthrough
	default:
		cache.db, err = bolt.New(bolt.OpenPath(config.StoragePath), bolt.Compression(bolt.GzipCompressor))
	}
	if err != nil {
		if config.Debug {
			log.WithFields(log.Fields{
				"config": config,
				"cache":  cache,
			}).Fatalf("boltgzipcache.New(): Open error: %v", err)
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
	if c.debug {
		log.WithFields(log.Fields{
			"bucketName": c.bucketName,
		}).Warnf("boltgzipcache.Close()")
	}
	return c.db.Close()
}

// Get retrieves the response corresponding to the given key if present.
func (c *Cache) Get(key string) (resp []byte, ok bool) {
	c.RLock()
	defer c.RUnlock()
	value, err := c.db.Get([]byte(c.bucketName), []byte(key))
	if c.debug {
		log.WithFields(log.Fields{
			"bucketName": c.bucketName,
			"key":        key,
			"ok":         resp != nil,
		}).Info("boltgzipcache.Get() OK")
	}
	return value, err != nil
}

// Set stores a response to the cache at the given key.
func (c *Cache) Set(key string, resp []byte) {
	c.Lock()
	defer c.Unlock()
	err := c.db.Put([]byte(c.bucketName), []byte(key), resp)
	if err != nil {
		if c.debug {
			log.WithFields(log.Fields{
				"bucketName": c.bucketName,
				"key":        key,
			}).Fatalln("boltgzipcache.Set() FAILURE", err)
		}
		return
	}
	if c.debug {
		log.WithFields(log.Fields{
			"bucketName": c.bucketName,
			"key":        key,
			"ok":         err != nil,
		}).Info("boltgzipcache.Set() OK")
	}
}

// Delete removes the response with the given key from the cache.
func (c *Cache) Delete(key string) {
	c.RLock()
	defer c.RUnlock()
	log.WithFields(log.Fields{
		"bucketName": c.bucketName,
		"key":        key,
	}).Warn("boltgzipcache.Delete() Not implemeted")
}
