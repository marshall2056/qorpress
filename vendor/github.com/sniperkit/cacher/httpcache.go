// Package httpcache provides a http.RoundTripper implementation that works as a
// mostly RFC-compliant cache for http responses.
//
// It is only suitable for use as a 'private' cache (i.e. for a web-browser or an API-client
// and not for a shared proxy).
//
package httpcache

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"strings"

	// "github.com/lox/httpcache"
	// "github.com/satoshun/go-http-cache"
	log "github.com/sirupsen/logrus"
)

const (
	stale = iota
	fresh
	transparent
	XFromCache         = "X-From-Cache" // XFromCache is the header added to responses that are returned from the cache
	defaultBufSize int = 8192           // 8192
)

var (
	cacheableResponseCodes = map[int]struct{}{
		http.StatusOK:                   {}, // 200
		http.StatusNonAuthoritativeInfo: {}, // 203
		http.StatusMultipleChoices:      {}, // 300
		http.StatusMovedPermanently:     {}, // 301
		http.StatusFound:                {}, // 302
		http.StatusNotFound:             {}, // 404
		http.StatusGone:                 {}, // 410
	}
	etagKeys           []string = []string{"Etag"}
	RequestBufferSize  int      = 8192
	ResponseBufferSize int      = 8192
	Verbose            bool
	Debug              bool
)

/*
	Refs:
	- https://github.com/nin-gen-sei/yisucon/blob/master/benchmarker/cache/cache.go
	- https://github.com/satoshun/go-http-cache/blob/master/registry.go
	-
*/

// A Cache interface is used by the Transport to store and retrieve responses.
type Cache interface {
	Get(key string) (responseBytes []byte, ok bool) // Get returns the []byte representation of a cached response and a bool set to true if the value isn't empty
	Set(key string, responseBytes []byte)           // Set stores the []byte representation of a response against a key
	Delete(key string)                              // Delete removes the value associated with the key
	// Upload(Digestable, io.Reader) error             // Upload content to cache
	// Contains(*pb.Digest) (bool, error)              // Check if cache contains
}

type cacheControl map[string]string

// cacheKey returns the cache key for req.
func cacheKey(req *http.Request) string {
	if req.Method == http.MethodGet {
		return req.URL.String()
	} else {
		return req.Method + " " + req.URL.String()
	}
}

// CachedResponse returns the cached http.Response for req if present, and nil
// otherwise.
func CachedResponse(c Cache, req *http.Request) (resp *http.Response, err error) {
	cachedVal, ok := c.Get(cacheKey(req))
	if !ok {
		return
	}
	b := bytes.NewBuffer(cachedVal)
	return http.ReadResponse(bufio.NewReader(b), req)
	/*
		b := bytes.NewReader(cachedVal)
		n := b.Len()
		if n > defaultBufSize {
			n = defaultBufSize
		}
		if Debug {
			log.WithFields(log.Fields{
				"BufSizer": n,
			}).Warn("httpcache.CachedResponse()")
		}
		return http.ReadResponse(bufio.NewReaderSize(b, n), req)
	*/
}

func newGatewayTimeoutResponse(req *http.Request) *http.Response {
	var braw bytes.Buffer
	braw.WriteString("HTTP/1.1 504 Gateway Timeout\r\n\r\n")
	resp, err := http.ReadResponse(bufio.NewReader(&braw), req)
	// braw := strings.NewReader("HTTP/1.1 504 Gateway Timeout\r\n\r\n")
	// resp, err := http.ReadResponse(bufio.NewReaderSize(braw, 0), req)
	if err != nil {
		log.WithFields(log.Fields{
			"resp": resp,
		}).Panicf("httpcache.newGatewayTimeoutResponse()", err)
		// panic(err)
	}
	return resp
}

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
// (This function copyright goauth2 authors: https://code.google.com/p/goauth2)
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header)
	for k, s := range r.Header {
		r2.Header[k] = s
		// r2.Header[k] = append([]string(nil), s...) // ref. ...?
	}
	if Debug {
		log.WithFields(log.Fields{
			"r2.Header": r2.Header,
		}).Info("httpcache.cloneRequest()")
	}
	return r2
}

func parseCacheControl(headers http.Header) cacheControl {
	cc := cacheControl{}
	ccHeader := headers.Get("Cache-Control")
	for _, part := range strings.Split(ccHeader, ",") {
		part = strings.Trim(part, " ")
		if part == "" {
			continue
		}
		if strings.ContainsRune(part, '=') {
			keyval := strings.Split(part, "=")
			cc[strings.Trim(keyval[0], " ")] = strings.Trim(keyval[1], ",")
		} else {
			cc[part] = ""
		}
	}
	if Debug {
		log.WithFields(log.Fields{
			"Cache-Control": cc,
		}).Warn("httpcache.parseCacheControl()")
	}
	return cc
}

// cachingReadCloser is a wrapper around ReadCloser R that calls OnEOF
// handler with a full copy of the content read from R when EOF is
// reached.
type cachingReadCloser struct {
	// Underlying ReadCloser.
	R io.ReadCloser
	// OnEOF is called with a copy of the content of R when EOF is reached.
	OnEOF func(io.Reader)
	// buf stores a copy of the content of R.
	buf bytes.Buffer
}

// Read reads the next len(p) bytes from R or until R is drained. The
// return value n is the number of bytes read. If R has no data to
// return, err is io.EOF and OnEOF is called with a full copy of what
// has been read so far.
func (r *cachingReadCloser) Read(p []byte) (n int, err error) {
	n, err = r.R.Read(p)
	r.buf.Write(p[:n])
	if err == io.EOF {
		r.OnEOF(bytes.NewReader(r.buf.Bytes()))
	}
	return n, err
}

func (r *cachingReadCloser) Close() error {
	return r.R.Close()
}
