// Package rediscache provides an implementation of httpcache.Cache that stores and
// retrieves data using Redis.
package rediscache

import (
	"log"
	"time"

	"gopkg.in/redis.v3"
	// "github.com/sniperkit/httpcache/helpers"
)

// Cache objects store and retrieve data using Redis.
type Cache struct {
	Client     *redis.Client
	Expiration time.Duration
}

// New returns a new Cache
func New(c *redis.Client, exp time.Duration) *Cache {
	return &Cache{Client: c, Expiration: exp}
}

func (c *Cache) Get(key string) (resp []byte, ok bool) {
	resp, err := c.Client.Get(key).Bytes()
	if err != nil {
		return []byte{}, false
	}
	return resp, true
}

func (c *Cache) Set(key string, content []byte) {
	err := c.Client.Set(key, content, c.Expiration).Err()
	if err != nil {
		log.Printf("Can't insert record in redis: %v\n", err)
	}
}

func (c *Cache) Delete(key string) {
	count, err := c.Client.Del(key).Result()
	if count == 0 || err != nil {
		log.Printf("Can't insert record in redis: %v\n", err)
	}
}

func (c *Cache) Indexes() {}
