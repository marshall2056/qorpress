# sniperkit-httpcache
=========

[WIP]

## Summary

Package httpcache provides a http.RoundTripper implementation that works as a mostly RFC-compliant cache for http responses.

It is only suitable for use as a 'private' cache (i.e. for a web-browser or an API-client and not for a shared proxy).

## Cache Backends
--------------

### Defaults
- The built-in 'memory' cache stores responses in an in-memory map.

### Memory
- [`plugins/lru`](https://github.com/sniperkit/httpcache/tree/master/plugins/lru) provides an in-memory cache that will evict least-recently used entries. (original: [`github.com/die-net/lrucache`](https://github.com/die-net/lrucache/tree/master/twotier))
- [`plugins/lru/twotier`](https://github.com/sniperkit/httpcache/tree/master/plugins/lru/twotier) allows caches to be combined, for example to use lrucache above with a persistent disk-cache. (original: [`github.com/die-net/lrucache/twotier`](https://github.com/die-net/lrucache/tree/master/twotier))

### File System - Storage

#### Local
- [`github.com/gregjones/httpcache/diskcache`](https://github.com/gregjones/httpcache/tree/master/diskcache) provides a filesystem-backed cache using the [diskv](https://github.com/peterbourgon/diskv) library.
- [`plugins/bbolt`](https://github.com/sniperkit/httpcache/tree/master/plugins/bbolt)
- [`plugins/boltdb-gzip`](https://github.com/sniperkit/httpcache/tree/master/plugins/boltdb-gzip)
- [`plugins/boltdb-ttl`](https://github.com/sniperkit/httpcache/tree/master/plugins/boltdb-ttl)
- [`plugins/storm`](https://github.com/sniperkit/httpcache/tree/master/plugins/storm)
- [`plugins/leveldb`](https://github.com/sniperkit/httpcache/tree/master/plugins/leveldb) provides a filesystem-backed cache using [leveldb](https://github.com/syndtr/goleveldb/leveldb).

#### Cloud
- [`plugins/azurestorage`](https://github.com/sniperkit/httpcache/tree/master/plugins/azurestorage) uses Azure Storage service. (original: [`github.com/PaulARoy/azurestoragecache`](https://github.com/PaulARoy/azurestoragecache))
- [`plugins/gcs`](https://github.com/sniperkit/httpcache/tree/master/plugins/gcs) uses Google cloud service engine. (original: [`github.com/PaulARoy/azurestoragecache`](https://github.com/PaulARoy/azurestoragecache))
- [`plugins/diskv/s3`](https://github.com/sniperkit/httpcache/tree/master/plugins/diskv/s3) uses Amazon S3 for storage. (original: [`github.com/sourcegraph/s3cache`](https://github.com/sourcegraph/s3cache))

### KV
- [`plugins/etcd/v2`](https://github.com/sniperkit/httpcache/tree/master/plugins/etcd/v2) provides etcd api v2 implentation
- [`plugins/etcd/v3`](https://github.com/sniperkit/httpcache/tree/master/plugins/etcd/v2) provides etcd api v3 implentation
- [`plugins/e3ch`](https://github.com/sniperkit/httpcache/tree/master/plugins/etcd/v2) provides etcd api v3 implentation with hierarchy
- [`plugins/memcache`](https://github.com/sniperkit/httpcache/tree/master/plugins/memcache) provides memcache implementations, for both App Engine and 'normal' memcache servers. (Original: [`memcache`](https://github.com/gregjones/httpcache/tree/master/memcache))

### RDB
- [`plugins/gorm`](https://github.com/sniperkit/httpcache/tree/master/plugins/gorm) provides gorm implementations, mainly for debugging and development

## Getting started

Below is a basic example of usage.
```
func httpCacheExample() {
    numOfRequests := 0
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=10"))
        if numOfRequests == 0 {
            w.Write([]byte("Hello!"))
        } else {
            w.Write([]byte("Goodbye!"))
        }
        numOfRequests++
    }))

    httpClient := &http.Client{
        Transport: httpcache.NewMemoryCacheTransport(),
    }
    makeRequest(ts, httpClient) // "Hello!"

    // The second request is under max-age, so the cache is used rather than hitting the server
    makeRequest(ts, httpClient) // "Hello!"

    // Sleep so the max-age is passed
    time.Sleep(time.Second * 11)

    makeRequest(ts, httpClient) // "Goodbye!"
}

func makeRequest(ts *httptest.Server, httpClient *http.Client) {
    resp, _ := httpClient.Get(ts.URL)
    var buf bytes.Buffer
    io.Copy(&buf, resp.Body)
    println(buf.String())
}
```

License
-------

-	[MIT License](LICENSE.txt)
