package limitedcache

// Package limitedcache provides an implementation of httpcache.Cache
// that limits the number of cached response written to disk files
// Content comes mainly from github.com/gregjones/httpcache

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"

	"log"

	"github.com/bluele/gcache"
	"github.com/peterbourgon/diskv"
	// "github.com/sniperkit/httpcache/helpers"
)

const maxEventChanLen = 1024

/*
	Refs:
	- https://github.com/wojtekzw/limitedcache/blob/master/example/example.go
*/

// Cache is an implementation of httpcache.Cache with persistent storage.
type Cache struct {
	d      *diskv.Diskv
	kc     gcache.Cache // key usage to help remove unused files from d
	eventC chan CacheOp
	//  lost messages counter in sending to channel eg. when no receivers
	lost int

	m sync.Mutex
}

// Get returns the response corresponding to key if present.
func (c *Cache) Get(key string) (resp []byte, ok bool) {
	f := keyToFilename(key)
	resp, err := c.d.Read(f)
	c.send(GetOp, key, f, err)
	if err != nil {
		return []byte{}, false
	}
	c.kc.Get(f)
	return resp, true
}

// Set saves a response to the cache as key.
func (c *Cache) Set(key string, resp []byte) {
	f := keyToFilename(key)
	err := c.d.WriteStream(f, bytes.NewReader(resp), true)
	c.send(SetOp, key, f, err)
	c.kc.Set(f, struct{}{})
}

// Delete removes the response with key from the cache.
func (c *Cache) Delete(key string) {
	f := keyToFilename(key)
	err := c.d.Erase(f)
	c.send(DeleteOp, key, f, err)
	c.kc.Remove(f)
}

// Events returns channel with cache operations messages.
func (c *Cache) Events() <-chan CacheOp {
	return c.eventC
}

// Lost returns number of lost messages - not sent to channel.
func (c *Cache) Lost() int {
	return c.lost
}

// ResetLost sets lost counter to 0.
func (c *Cache) ResetLost() int {
	c.m.Lock()
	defer c.m.Unlock()
	c.lost = 0
	return c.lost
}

// LoadKeysFromDisk - loads cache keys to memory to keep limited cache
func (c *Cache) LoadKeysFromDisk(basePath string) {
	err := filepath.Walk(basePath, func(path string, f os.FileInfo, err error) error {
		if err == nil && !f.IsDir() {
			c.kc.Set(filepath.Base(path), struct{}{})
		}
		return err
	})

	if err != nil {
		log.Printf("error loading keys from disk: %v,", err)
	}
	log.Printf("loaded keys from disk: %d", c.kc.Len())

}

func (c *Cache) send(op OpType, key, file string, err error) {
	c.m.Lock()
	defer c.m.Unlock()

	if len(c.eventC) == maxEventChanLen {
		c.lost++
		return
	}
	c.eventC <- CacheOp{op: op, key: key, file: file, err: err}
}

func keyToFilename(key string) string {
	h := md5.New()
	io.WriteString(h, key)
	s := hex.EncodeToString(h.Sum(nil))
	return s
}

// New returns a new Cache that will store files in basePath
func New(basePath string, limit int) *Cache {

	d := diskv.New(diskv.Options{
		BasePath:     path.Join(basePath, ""),
		Transform:    func(s string) []string { return []string{s[0:2], s[2:4]} },
		CacheSizeMax: 100 * 1024 * 1024, // 100MB
	})

	kc := gcache.New(limit).LFU().EvictedFunc(func(key, value interface{}) {
		d.Erase(key.(string))
	}).Build()

	return &Cache{
		d:      d,
		kc:     kc,
		eventC: make(chan CacheOp, maxEventChanLen),
	}
}

// NewWithDiskv returns a new Cache using the provided Diskv as underlying
// storage.
func NewWithDiskv(d *diskv.Diskv, limit int) *Cache {
	kc := gcache.New(limit).LFU().EvictedFunc(func(key, value interface{}) {
		d.Erase(key.(string))
	}).Build()

	return &Cache{
		d:      d,
		kc:     kc,
		eventC: make(chan CacheOp, maxEventChanLen),
	}
}
