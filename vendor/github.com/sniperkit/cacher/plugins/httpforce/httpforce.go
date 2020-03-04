package httpforcecache

import (
	"net/http"
	"net/http/httputil"

	"github.com/sniperkit/httpcache"
)

// cacheKey returns the cache key for req.
func cacheKey(req *http.Request) string {
	return req.URL.String()
}

type Transport struct {
	Transport http.RoundTripper
	Cache     httpcache.Cache
}

func NewTransport(c httpcache.Cache) *Transport {
	return &Transport{Cache: c}
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	cacheKey := cacheKey(req)

	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	cachedResp, err := httpcache.CachedResponse(t.Cache, req)
	if cachedResp != nil && err == nil {
		return cachedResp, nil
	}
	if cachedResp == nil && err == nil {
		resp, err = transport.RoundTrip(req)
		if cachableResponse(resp) {
			respBytes, err := httputil.DumpResponse(resp, true)
			if err == nil {
				t.Cache.Set(cacheKey, respBytes)
			}
		}
	}

	return resp, nil
}

func cachableResponse(resp *http.Response) bool {
	code := resp.StatusCode
	if 200 <= code && code < 300 {
		return true
	}
	return false
}

func NewMemoryCacheTransport() *Transport {
	c := httpcache.NewMemoryCache()
	t := NewTransport(c)
	return t
}

func (t *Transport) DeleteCache(req *http.Request) {
	key := cacheKey(req)
	t.Cache.Delete(key)
}
