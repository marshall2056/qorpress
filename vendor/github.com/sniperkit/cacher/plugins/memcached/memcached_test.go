package memcached_cache

import (
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) Test(c *C) {
	cache := New("localhost:11211", time.Minute*10)
	cache.Indexes()

	key := "testKey"
	_, ok := cache.Get(key)

	c.Assert(ok, Equals, false)

	val := []byte("some bytes")
	cache.Set(key, val)

	retVal, ok := cache.Get(key)
	c.Assert(ok, Equals, true)
	c.Assert(string(retVal), Equals, string(val))

	val = []byte("some other bytes")
	cache.Set(key, val)

	retVal, ok = cache.Get(key)
	c.Assert(ok, Equals, true)
	c.Assert(string(retVal), Equals, string(val))

	cache.Delete(key)

	_, ok = cache.Get(key)
	c.Assert(ok, Equals, false)
}
