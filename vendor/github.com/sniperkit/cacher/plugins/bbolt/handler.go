//
package bboltdbcache

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	bolt "github.com/coreos/bbolt"
	log "github.com/sirupsen/logrus"
	// "github.com/sniperkit/httpcache/helpers"
)

/*
	Refs:
	- https://github.com/br0xen/boltbrowser
	- https://github.com/aerth/fforum/blob/master/forum.go
	- https://github.com/Everlag/poeitemstore/blob/master/stash/stash.go (encoding json)
*/

const (
	bktName                     string = "httpcache"
	bktDir                      string = "./shared/data/cache/.bbolt"
	defaultStorageFileExtension string = ".bbolt"
)

var (
	defaultStorageDir  string = filepath.Join(bktDir, bktName)
	defaultStorageFile string = fmt.Sprintf("%s/%s%s", defaultStorageDir, bktName, defaultStorageFileExtension)
)

// Cache is an implementation of httpcache.Cache that uses a bolt database.
type Cache struct {
	// sync.Mutex
	sync.RWMutex
	db          *bolt.DB
	bucket      *bolt.Bucket
	storagePath string
	bucketName  string
	debug       bool
}

type Config struct {
	BucketName     string
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
		}).Warnf("bboltcache.New()")
	}

	if config == nil {
		log.Warnln("config is nil")
		config.StoragePath = defaultStorageFile
		config.BucketName = bktName
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
		}).Warnf("bboltcache.New() ---> post-processed")
	}

	var err error
	cache := &Cache{}
	cache.storagePath = config.StoragePath
	cache.bucketName = config.BucketName
	cache.debug = config.Debug

	cache.db, err = bolt.Open(config.StoragePath, 0777, nil)
	if err != nil {
		if config.Debug {
			log.WithFields(log.Fields{
				"config": config,
				"cache":  cache,
			}).Fatalf("bboltcache.New(): Open error: %v", err)
		}
		return nil, err
	}

	init := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(config.BucketName))
		return err
	}

	if err := cache.db.Update(init); err != nil {
		if config.Debug {
			log.Fatalf("bboltcache.New(): init error: %v", err)
		}
		if err := cache.db.Close(); err != nil {
			if config.Debug {
				log.Fatalf("bboltcache.New(): close error: %v", err)
			}
		}
		return nil, err
	}
	return cache, nil
}

// Mount returns a new Cache using the provided (and opened) bolt database.
func Mount(db *bolt.DB) *Cache {
	// log.Info("bboltcache.Mount() [CONNECT] ===> BoltDB")
	return &Cache{db: db}
}

// Close closes the underlying boltdb database.
func (c *Cache) Close() error {
	if c.debug {
		log.WithFields(log.Fields{
			"bucketName": c.bucketName,
		}).Warnf("bboltcache.Close()")
	}
	return c.db.Close()
}

// Get retrieves the response corresponding to the given key if present.
func (c *Cache) Get(key string) (resp []byte, ok bool) {
	get := func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(c.bucketName))
		if bkt == nil {
			return errors.New("bucket is nil")
		}
		resp = bkt.Get([]byte(key))
		return nil
	}

	if err := c.db.View(get); err != nil {
		log.Printf("boltdbcache.Get(): view error: %v", err)
		return resp, false
	}
	/*
		// ref. https://github.com/jonasi/ghsync/blob/master/data/boltdb/repositories.go
		var dec interface{}
		if err := decodeVal(resp, &dec); err != nil {
			return err
		}
		return dec, dec != nil
	*/
	return resp, resp != nil
}

// Set stores a response to the cache at the given key.
func (c *Cache) Set(key string, resp []byte) {
	set := func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(c.bucketName))
		if bkt == nil {
			return errors.New("bucket is nil")
		}
		/*
			enc, err := encodeVal(resp)
			if err != nil {
				return err
			}

			return bkt.Put([]byte(key), enc)
		*/
		return bkt.Put([]byte(key), resp)
	}

	if err := c.db.Update(set); err != nil {
		log.Printf("boltdbcache.Set(): update error: %v", err)
	}
}

func encodeInt(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func decodeInt(b []byte) int {
	return int(binary.BigEndian.Uint64(b))
}

func encodeVal(v interface{}) ([]byte, error) {
	var b bytes.Buffer

	if err := gob.NewEncoder(&b).Encode(v); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func decodeVal(b []byte, dest interface{}) error {
	return gob.NewDecoder(bytes.NewReader(b)).Decode(dest)
}

// Delete removes the response with the given key from the cache.
func (c *Cache) Delete(key string) {
	// c.RLock()
	// defer c.RUnlock()
	del := func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(c.bucketName))
		if bkt == nil {
			if c.debug {
				log.WithFields(log.Fields{
					"context":    "Bucket",
					"bucketName": c.bucketName,
					"key":        key,
				}).Fatal("bboltcache.Delete() Bucket error")
			}
			return errors.New(fmt.Sprintf("bboltcache.Delete(): could not reach the bucket: %s", c.bucketName))
		}
		return bkt.Delete([]byte(key))
	}
	if err := c.db.Update(del); err != nil {
		if c.debug {
			log.WithFields(log.Fields{
				"context":    "Delete",
				"bucketName": c.bucketName,
				"key":        key,
			}).Fatalln("bboltcache.Delete() Update: ", err)
		}
		return
	}
	if c.debug {
		log.WithFields(log.Fields{
			"context":    "Update",
			"bucketName": c.bucketName,
			"key":        key,
		}).Info("bboltcache.Delete() OK")
	}
}

func (c *Cache) findAll() map[string][]byte {
	m := make(map[string][]byte)
	c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("default"))
		if b == nil {
			return errors.New("can't get [default] bucket in boltdb")
		}
		b.ForEach(func(k, v []byte) error {
			if c.debug {
				log.Println(string(k))
			}
			// must copy key, value to out varialbe (out of this transaction scope)
			// to use the value, or creat a new one and use it immediately
			if len(v) == 0 {
				m[string(k)] = nil
			} else {
				value := make([]byte, len(v))
				copy(value, v)
				m[string(k)] = value
			}
			return nil
		})

		return nil
	})
	return m
}

// EmptyBoltBucket -- empty a bucket if not exist, create one
func (c *Cache) truncate() error {
	err := c.db.Update(func(tx *bolt.Tx) error {
		namebytes := []byte(c.bucketName)
		bucket := tx.Bucket(namebytes)
		if bucket != nil {
			err := tx.DeleteBucket(namebytes)
			if err != nil {
				if c.debug {
					log.Printf("empty bucket - [%s] error\n", c.bucketName)
				}
				return err
			}
		}
		_, err := tx.CreateBucket(namebytes)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}
