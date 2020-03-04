package diskvcache

import (
	"bytes"
	"github.com/dustin/go-humanize"
	"io/ioutil"
	"os"
	"testing"
	//"github.com/cloudfoundry/bytefmt"
)

//bytefmt.ByteSize(100.5*bytefmt.MEGABYTE) // returns "100.5M"
//bytefmt.ByteSize(uint64(1024)) // returns "1K"

func TestDiskCache(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "httpcache")
	if err != nil {
		t.Fatalf("TempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cache := New(tempDir)

	// x ->
	// 100 ->

	t.Logf("cache expire: %#d, %#s", cache.GetCacheSizeMax(), humanize.Bytes(cache.GetCacheSizeMax()))

	var cacheTime uint64 = 400 * 1024 * 1024
	cache.SetNewCacheSizeMax(cacheTime)
	t.Logf("cache expire: %#d, %#s", cache.GetCacheSizeMax(), humanize.Bytes(cache.GetCacheSizeMax()))

	key := "testKey"
	_, ok := cache.Get(key)
	if ok {
		t.Fatal("retrieved key before adding it")
	}

	val := []byte("some bytes")
	cache.Set(key, val)

	retVal, ok := cache.Get(key)
	if !ok {
		t.Fatal("could not retrieve an element we just added")
	}
	if !bytes.Equal(retVal, val) {
		t.Fatal("retrieved a different value than what we put in")
	}

	cache.Delete(key)

	_, ok = cache.Get(key)
	if ok {
		t.Fatal("deleted key still present")
	}
}
