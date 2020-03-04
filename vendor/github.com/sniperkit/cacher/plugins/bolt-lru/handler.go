// This package provides a simple LRU cache backed by boltdb It is based on the
// LRU implementation in groupcache:
// https://github.com/golang/groupcache/tree/master/lru
//
package boltlrucache

import (
	"container/list"
	"errors"
	"sync"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

const (
	bktName                     string = "httpcache"
	bktDir                      string = "./shared/data/cache/.bolt-lru"
	defaultStorageFileExtension string = ".lru.boltdb"
	defaultCacheSize            int    = 128 * 1024 * 1024 // 128MB cache
)

var (
	defaultStorageDir  string = filepath.Join(bktDir, bktName)
	defaultStorageFile string = fmt.Sprintf("%s/%s%s", defaultStorageDir, bktName, defaultStorageFileExtension)
)

// var bucketName = []byte("cache")

type Cache struct {
	size        int
	evictList   *list.List
	items       map[string]*list.Element
	lock        sync.RWMutex
	db          *bolt.DB
	stop        chan bool
	storagePath string
	bucketName  string
	debug       bool
}

type Config struct {
	BucketName     string
	CacheSize      int
	StoragePath    string
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
		}).Warnf("boltlrucache.New()")
	}

	if config == nil {
		log.Warnln("config is nil")
		config.StoragePath = defaultStorageFile
		config.BucketName = bktName
		config.CacheSize = defaultCacheSize
	}

	if config.CacheSize <= 0 {
		config.CacheSize = defaultCacheSize
		/*
			if config.StrictMode {
				return nil, errors.New("Must provide a positive size")
			}
		*/
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
		}).Warnf("boltlrucache.New() ---> post-processed")
	}

	db, err := bolt.Open(config.StoragePath, 0640, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists(config.BucketName)
		return err
	})
	if err != nil {
		return nil, err
	}

	c := &Cache{
		db:          db,
		evictList:   list.New(),
		items:       make(map[string]*list.Element, size),
		size:        config.CacheSize,
		storagePath: config.StoragePath,
		debug:       config.Debug,
		bucketName:  config.BucketName,
		storagePath: config.StoragePath,
	}

	c.stop = make(chan bool, 1)

	go func() {
		for {
			select {
			case <-c.stop:
				return
			}
		}
	}()

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(config.BucketName)
		cur := b.Cursor()
		var key string
		for k, _ := cur.First(); k != nil; k, _ = cur.Next() {
			key = string(k)
			el := c.evictList.PushFront(key)
			c.items[key] = el
		}
		return nil
	})
	return c, err
}

// Mount returns a new Cache using the provided (and opened) bolt database.
func Mount(db *bolt.DB, size int) *Cache {
	return &Cache{
		db:   db,
		size: size,
		stop: make(chan bool, 1),
	}
}

func (c *Cache) Add(key string, value []byte) error {
	return c.AddMulti(map[string][]byte{key: value})
}

func (c *Cache) AddMulti(data map[string][]byte) error {
	keysToRemove := make([]string, 0)
	c.lock.Lock()
	for key, _ := range data {
		if el, ok := c.items[key]; ok {
			c.evictList.MoveToFront(el)
		} else {
			el := c.evictList.PushFront(key)
			c.items[key] = el
			evict := c.evictList.Len() > c.size
			if evict {
				el := c.evictList.Back()
				if el != nil {
					key := el.Value.(string)
					c.evictList.Remove(el)
					delete(c.items, key)
					keysToRemove = append(keysToRemove, key)
				}
			}
		}
	}
	c.lock.Unlock()
	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.bucketName))
		var err error
		for _, k := range keysToRemove {
			err = b.Delete([]byte(k))
			if err != nil {
				if c.debug {
					log.WithFields(log.Fields{
						"bucketName": c.bucketName,
						"key":        k,
						"keysToRemove.Length": len(keysToRemove),
					}).Fatalf("boltlrucache.AddMulti().Delete ERROR: ", err)
				}
				return err
			}
		}

		for key, value := range data {
			err = b.Put([]byte(key), value)
			if err != nil {
				if c.debug && err != nil {
					log.WithFields(log.Fields{
						"bucketName": c.bucketName,
						"key":        key,
						"value":      value,
					}).Fatalf("boltlrucache.AddMulti().Put ERROR: ", err)
				}
				return err
			}
		}
		return nil
	})
}

func (c *Cache) Get(key string) (value []byte, err error) {
	data, err := c.MultiGet([]string{key})
	if err != nil {
		if c.debug {
			log.WithFields(log.Fields{
				"bucketName": c.bucketName,
				"key":        key,
			}).Fatalf("boltlrucache.Get().MultiGet ERROR: ", err)
		}
		return
	}
	value, ok := data[key]
	if !ok {
		if c.debug {
			log.WithFields(log.Fields{
				"bucketName": c.bucketName,
				"key":        key,
				"data":       data,
			}).Fatalf("boltlrucache.Get().MultiGet ERROR: ", err)
		}
		err = errors.New("not found")
	}
	return
}

func (c *Cache) MultiGet(keys []string) (values map[string][]byte, err error) {
	c.lock.Lock()
	existsKeys := make([]string, 0, len(keys))
	for _, k := range keys {
		if el, ok := c.items[k]; ok {
			existsKeys = append(existsKeys, k)
			c.evictList.MoveToFront(el)
		}
	}
	c.lock.Unlock()
	values = make(map[string][]byte, len(existsKeys))
	err = c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.bucketName))
		for _, k := range existsKeys {
			values[k] = b.Get([]byte(k))
		}
		return nil
	})
	return
}

func (c *Cache) Len() int {
	c.lock.RLock()
	length := c.evictList.Len()
	c.lock.RUnlock()
	// defer c.lock.RUnlock() // overhead with defer ?!
	return length
	// return c.evictList.Len()
}

func (c *Cache) Close() {
	// c.lock.Lock()
	c.stop <- true
	// defer c.Unlock()
	c.db.Close()
}
