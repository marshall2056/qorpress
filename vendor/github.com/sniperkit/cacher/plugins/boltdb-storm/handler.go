package boltstormcache

import (
	"fmt"
	"path/filepath"

	"github.com/asdine/storm"
	log "github.com/sirupsen/logrus"
	// "github.com/asdine/storm/codec/json"
)

const (
	bktName                     string = "httpcache"
	bktDir                      string = "./shared/data/cache/.storm"
	defaultStorageFileExtension string = ".storm"
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
	db          *storm.DB
	storagePath string
	bucketName  string
	debug       bool
}

type Config struct {
	BucketName  string
	StoragePath string
	Debug       bool
}

// New returns a new Cache that uses a bolt database at the given path.
func New(config *Config) (*Cache, error) {

	if config == nil {
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
		}).Warnf("boltstormcache.New()")
	}

	var err error
	cache := &Cache{}
	cache.storagePath = config.StoragePath
	cache.bucketName = config.BucketName
	cache.debug = config.Debug

	/*
		// ref. https://github.com/thesyncim/365a/blob/master/server/app.go
		db, err :=storm.Open("my.db",storm.AutoIncrement(),storm.Codec(json.Codec))
		if err != nil {
			return err
		}
	*/

	cache.db, err = storm.Open(config.StoragePath)
	if err != nil {
		return nil, err
	}
	defer cache.db.Close()

	return cache, nil
}

/*
// ref. https://github.com/thesyncim/365a/blob/master/server/module/module.go
func getExtraFields(id string) ([]Field, error) {
	var moduleInfo ModuleConfig
	err := stormdb.One("Id", id, &moduleInfo)
	return moduleInfo.Fields, err
}
*/

// Mount returns a new Cache using the provided (and opened) bolt database.
func Mount(db *storm.DB) *Cache {
	return &Cache{db: db}
}

// Close closes the underlying boltdb database.
func (c *Cache) Close() error {
	return c.db.Close()
}

// Get retrieves the response corresponding to the given key if present.
func (c *Cache) Get(key string) (resp []byte, ok bool) {
	return resp, resp != nil
}

// Set stores a response to the cache at the given key.
func (c *Cache) Set(key string, resp []byte) {
}

// Delete removes the response with the given key from the cache.
func (c *Cache) Delete(key string) {

}
