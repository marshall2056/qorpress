package limitedcache

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestDiskCache(t *testing.T) {
	cacheLimit := 10
	tempDir, err := ioutil.TempDir("", "httpcache")
	if err != nil {
		t.Fatalf("TempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cache := New(tempDir, cacheLimit)

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

	val = []byte("some bytes")
	for i := 0; i < 2*cacheLimit; i++ {
		cache.Set(fmt.Sprintf("some key:%d", i), val)
	}

	count, err := countFiles(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	if count != cacheLimit {
		t.Fatalf("expected %d files, got: %d", cacheLimit, count)
	}
}

func countFiles(basePath string) (int, error) {
	var count int
	err := filepath.Walk(basePath, func(path string, f os.FileInfo, err error) error {
		if err == nil && !f.IsDir() {
			count++
		}
		return err
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}
