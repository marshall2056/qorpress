package rediscache

import (
	"testing"
	"time"

	"gopkg.in/redis.v3"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) Test(c *C) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	cache := New(client, time.Second/3)
	cache.Indexes()

	key := "testKey"
	_, ok := cache.Get(key)

	c.Assert(ok, Equals, false)

	val := []byte("some bytes")
	cache.Set(key, val)

	retVal, ok := cache.Get(key)
	c.Assert(ok, Equals, true)
	c.Assert(string(retVal), Equals, string(val))

	time.Sleep(time.Second / 2)
	retVal, ok = cache.Get(key)
	c.Assert(ok, Equals, false)
	c.Assert(string(retVal), Equals, "")

	val = []byte("some other bytes")
	cache.Set(key, val)

	retVal, ok = cache.Get(key)
	c.Assert(ok, Equals, true)
	c.Assert(string(retVal), Equals, string(val))

	cache.Delete(key)

	_, ok = cache.Get(key)
	c.Assert(ok, Equals, false)
}
