// Package diskcache provides an implementation of httpcache.Cache that uses the diskv package
// to supplement an in-memory map with persistent storage
//
package diskvcache

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"

	"github.com/peterbourgon/diskv"
	log "github.com/sirupsen/logrus"
)

const (
	cacheFolder = "./shared/data/cache/.diskv"
)

var (
	defaultCacheSize uint64 = 256 * 1024 * 1024 // 256MB disk cache
)

// Cache is an implementation of httpcache.Cache that supplements the in-memory map with persistent storage
type Cache struct {
	store *diskv.Diskv
	debug bool
}

type Config struct {
	BasePath     string
	CacheSizeMax uint64
	Transform    string
	Debug        bool
}

// New returns a new Cache that will store files in basePath
func New(config *Config) *Cache {
	diskvConfig := diskv.Options{}
	if config == nil {
		diskvConfig.BasePath = cacheFolder
		diskvConfig.CacheSizeMax = defaultCacheSize
		// diskvConfig.Transform = getStorePath
	}

	if config.BasePath == "" {
		diskvConfig.BasePath = cacheFolder
	}

	if config.CacheSizeMax <= 0 {
		diskvConfig.CacheSizeMax = defaultCacheSize
	}

	/*
		if config.Transform == "" {
			diskvConfig.Transform = getStorePath
		}
	*/
	return &Cache{
		store: diskv.New(diskvConfig),
		debug: config.Debug,
	}
}

// NewWithDiskv returns a new Cache using the provided Diskv as underlying storage.
func Mount(store *diskv.Diskv) *Cache {
	return &Cache{store: store}
}

// Get returns the response corresponding to key if present
func (c *Cache) Get(key string) (resp []byte, ok bool) {
	key = keyToFilename(key)
	resp, err := c.store.Read(key)
	if err != nil {
		if !os.IsNotExist(err) {
			logError(err, key, "disk_cache_read", "bytes", "Failed to read bytes from disk cache")
		}
		return nil, false
	}
	return resp, true
}

// Set saves a response to the cache as key
func (c *Cache) Set(key string, resp []byte) {
	key = keyToFilename(key)
	// c.d.WriteStream(key, bytes.NewReader(resp), true)
	err := c.store.WriteStream(key, bytes.NewReader(resp), true)
	if err != nil {
		logError(err, key, "disk_cache_write", "bytes", "Failed to write bytes to disk cache")
	}
}

// Delete removes the response with key from the cache
func (c *Cache) Delete(key string) {
	key = keyToFilename(key)
	// c.store.Erase(key)
	err := c.store.Erase(key)
	if err != nil {
		logError(err, key, "disk_cache_delete", "bytes", "Failed to delete from disk cache")
	}
}

func (c *Cache) Flush() {
	c.store.EraseAll()
}

func (c *Cache) GetCacheSizeMax() uint64 {
	return c.store.CacheSizeMax
}

func (c *Cache) SetNewCacheSizeMax(expire uint64) {
	c.store.CacheSizeMax = expire
}

func keyToFilename(key string) string {
	h := md5.New()
	io.WriteString(h, key)
	return hex.EncodeToString(h.Sum(nil))
}

func getStorePath(key string) []string {
	const folderCharCount = 4
	folders := []string{}
	for len(key) > folderCharCount {
		folders = append(folders, key[:folderCharCount])
		key = key[folderCharCount+1:]
	}
	return folders
}

func md5sum(data []byte) string {
	md5sum := md5.Sum(data)
	return hex.EncodeToString(md5sum[:])
}

func transformKey(key string) string {
	// Transform all keys to an MD5 hash to guarantee valid file names
	return md5sum([]byte(key))
}

func logError(err error, key, code, category, msg string) {
	log.WithFields(log.Fields{
		"key":      key,
		"code":     code,
		"category": category,
		"msg":      msg,
	}).Fatalln("diskvcache.logError() View: ", err)
}
