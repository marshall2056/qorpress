package httpcache

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/sethgrid/pester"
	log "github.com/sirupsen/logrus"
)

/*
	Refs:
	- https://github.com/mreiferson/go-httpclient
		- Snippets:
		```Go
		transport := &httpclient.Transport{
		    ConnectTimeout:        1*time.Second,
		    RequestTimeout:        10*time.Second,
		    ResponseHeaderTimeout: 5*time.Second,
		}
		defer transport.Close()

		client := &http.Client{Transport: transport}
		req, _ := http.NewRequest("GET", "http://127.0.0.1/test", nil)
		resp, err := client.Do(req)
		if err != nil {
		    return err
		}
		defer resp.Body.Close()
		# Note: you will want to re-use a single client object rather than creating one for each request, otherwise you will end up leaking connections.
		```
	- https://github.com/heatxsink/go-httprequest
*/

// Transport is an implementation of http.RoundTripper that will return values from a cache
// where possible (avoiding a network request) and will additionally add validators (etag/if-modified-since)
// to repeated requests allowing servers to return 304 / Not Modified
type Transport struct {
	// The RoundTripper interface actually used to make requests
	// If nil, http.DefaultTransport is used
	Transport http.RoundTripper
	Cache     Cache
	// If true, responses returned from the cache will be given an extra header, X-From-Cache
	MarkCachedResponses bool
	Debug               bool
	processedCount      int
	cachedCount         int
	conditionalCount    int
	// transport *httputil.StatsTransport
}

// NewTransport returns a new Transport with the
// provided Cache implementation and MarkCachedResponses set to true
func NewTransport(c Cache) *Transport {
	return &Transport{Cache: c, MarkCachedResponses: true}
}

// Client returns an *http.Client that caches responses.
func (t *Transport) Client() *http.Client {
	return &http.Client{Transport: t}
}

func (t *Transport) Info() {
	log.WithFields(log.Fields{
		// "Transport.Debug":               t.Debug,
		"Transport.MarkCachedResponses": t.MarkCachedResponses,
		"Transport.processedCount":      t.processedCount,
		"Transport.cachedCount":         t.cachedCount,
		"Transport.conditionalCount":    t.conditionalCount,
	}).Info("httpcache.Info()")
}

func (t *BlockingTransport) transport() http.RoundTripper {
	if t.Transport == nil {
		return http.DefaultTransport
	}
	return t.Transport
}

/*
	Refs:
	- https://github.com/evepraisal/go-evepraisal/blob/master/evepraisal/app.go#L48-L71
*/
// control the resiliency
func (t *Transport) Pester() *pester.Client {
	client := pester.New()
	client.Concurrency = 3
	client.MaxRetries = 5
	client.Backoff = pester.ExponentialBackoff
	client.KeepLog = true
	return client
}

// RoundTrip takes a Request and returns a Response
//
// If there is a fresh Response already in cache, then it will be returned without connecting to
// the server.
//
// If there is a stale Response, then any validators it contains will be set on the new request
// to give the server a chance to respond with NotModified. If this happens, then the cached Response
// will be returned.
func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	t.processedCount++
	cacheKey := cacheKey(req)
	// cacheable := (req.Method == "GET" || req.Method == "POST" || req.Method == "HEAD") && req.Header.Get("range") == ""
	cacheable := (req.Method == "GET" || req.Method == "HEAD") && req.Header.Get("Range") == ""
	var cachedResp *http.Response
	if cacheable {
		cachedResp, err = CachedResponse(t.Cache, req)
	} else {
		// Need to invalidate an existing value
		t.Cache.Delete(cacheKey)
	}
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	if cacheable && cachedResp != nil && err == nil {
		if t.MarkCachedResponses {
			cachedResp.Header.Set(XFromCache, "1")
			if t.Debug {
				log.WithFields(log.Fields{
					"cacheKey":     cacheKey,
					"X-From-Cache": req.Header.Get(XFromCache),
				}).Debug("httpcache.RoundTrip() MarkCachedResponses")
			}
		}

		if varyMatches(cachedResp, req) {
			// Can only use cached value if the new request doesn't Vary significantly
			freshness := getFreshness(cachedResp.Header, req.Header)
			if freshness == fresh {
				if t.Debug {
					log.WithFields(log.Fields{
						"cacheKey":  cacheKey,
						"freshness": freshness,
					}).Debug("httpcache.RoundTrip() getFreshness")
				}
				return cachedResp, nil
			}

			if freshness == stale {
				var req2 *http.Request
				// Add validators if caller hasn't already done so
				var etagKey, etag string
				for _, v := range etagKeys {
					etag = cachedResp.Header.Get(v)
					if etag != "" {
						etagKey = v
						if t.Debug {
							log.WithFields(log.Fields{
								"cacheKey":        cacheKey,
								"etagKey":         etagKey,
								"etag":            etag,
								"req.Header.etag": req.Header.Get(etagKey),
							}).Warn("httpcache.RoundTrip() ETAG Match.")
						}
						break
					}
				}

				// etag := cachedResp.Header.Get(etagKey)
				if etag != "" && req.Header.Get(etagKey) == "" {
					req2 = cloneRequest(req)
					req2.Header.Set("If-None-Match", etag)
					if t.Debug {
						log.WithFields(log.Fields{
							"cacheKey": cacheKey,
							// "etag":                      etag,
							// "etagKey":                   etagKey,
							// "req.Header.etag":           req.Header.Get(etagKey),
							// "req2.Header.If-None-Match": req2.Header.Get("If-None-Match"),
						}).Warn("httpcache.RoundTrip() cloneRequest, etag not empty in cachedResp, but header request was empty.")
					}
				}
				lastModified := cachedResp.Header.Get("Last-Modified")
				if lastModified != "" && req.Header.Get("Last-Modified") == "" {
					if req2 == nil {
						if t.Debug {
							log.WithFields(log.Fields{
								"cacheKey": cacheKey,
								// "etag":            etag,
								// "lastModified":    lastModified,
								// "etagKey":         etagKey,
								// "req.Header.etag": req.Header.Get(etagKey),
							}).Warn("httpcache.RoundTrip() cloneRequest, cloning request as it was nil.")
						}
						req2 = cloneRequest(req)
					}
					req2.Header.Set("If-Modified-Since", lastModified)
				}
				if req2 != nil {
					req = req2
					if t.Debug {
						log.Warn("httpcache.RoundTrip() cloneRequest success...")
					}
				}
				if t.Debug {
					log.WithFields(log.Fields{
						"cacheKey":        cacheKey,
						"freshness":       freshness,
						"etag":            etag,
						"etagKey":         etagKey,
						"req.Header.etag": req.Header.Get(etagKey),
						"lastModified":    lastModified,
					}).Warn("httpcache.RoundTrip() ETAG")
				}
			}
		}

		resp, err = transport.RoundTrip(req)

		if t.Debug {
			log.WithFields(log.Fields{
				"resp.StatusCode":        resp.StatusCode,
				"resp.Header.Length":     len(resp.Header),
				"req.Method":             req.Method,
				"http.StatusNotModified": http.StatusNotModified,
				"cacheKey":               cacheKey,
			}).Info("httpcache.RoundTrip() transport.RoundTrip")
		}

		if err == nil && req.Method == "GET" && resp.StatusCode == http.StatusNotModified {
			// Replace the 304 response with the one from cache, but update with some new headers
			endToEndHeaders := getEndToEndHeaders(resp.Header)
			for _, header := range endToEndHeaders {
				cachedResp.Header[header] = resp.Header[header]
			}
			// cachedResp.Status = fmt.Sprintf("%d %s", http.StatusOK, http.StatusText(http.StatusOK))
			// cachedResp.StatusCode = http.StatusOK
			resp = cachedResp
		} else if (err != nil || (cachedResp != nil && resp.StatusCode >= 500)) &&
			req.Method == "GET" && canStaleOnError(cachedResp.Header, req.Header) {
			// In case of transport failure and stale-if-error activated, returns cached content
			// when available
			// cachedResp.Status = fmt.Sprintf("%d %s", http.StatusOK, http.StatusText(http.StatusOK))
			// cachedResp.StatusCode = http.StatusOK
			return cachedResp, nil
		} else {
			if err != nil || resp.StatusCode != http.StatusOK {
				t.Cache.Delete(cacheKey)
			}
			if err != nil {
				return nil, err
			}
		}
	} else {
		reqCacheControl := parseCacheControl(req.Header)
		if t.Debug {
			log.WithFields(log.Fields{
				"cacheable":       cacheable,
				"cacheKey":        cacheKey,
				"reqCacheControl": reqCacheControl,
			}).Info("httpcache.RoundTrip() reqCacheControl")
		}
		if _, ok := reqCacheControl["Only-If-Cached"]; ok {
			resp = newGatewayTimeoutResponse(req)
		} else {
			resp, err = transport.RoundTrip(req)
			if err != nil {
				return nil, err
			}
		}
	}

	if cacheable && canStore(resp.StatusCode, parseCacheControl(req.Header), parseCacheControl(resp.Header)) {
		// if cacheable && canStore(parseCacheControl(req.Header), parseCacheControl(resp.Header)) {
		for _, varyKey := range headerAllCommaSepValues(resp.Header, "Vary") {
			varyKey = http.CanonicalHeaderKey(varyKey)
			fakeHeader := "X-Varied-" + varyKey
			reqValue := req.Header.Get(varyKey)
			if reqValue != "" {
				resp.Header.Set(fakeHeader, reqValue)
			}
		}
		if t.Debug {
			log.WithFields(log.Fields{
				"cacheKey":        cacheKey,
				"req.Method":      req.Method,
				"resp.StatusCode": resp.StatusCode,
			}).Warn("httpcache.RoundTrip() cachingReadCloser")
		}
		switch req.Method {
		case "GET":
			// Delay caching until EOF is reached.
			resp.Body = &cachingReadCloser{
				R: resp.Body,
				OnEOF: func(r io.Reader) {
					resp := *resp
					resp.Body = ioutil.NopCloser(r)
					respBytes, err := httputil.DumpResponse(&resp, true)
					if err == nil {
						t.Cache.Set(cacheKey, respBytes)
					}
				},
			}
		default:
			respBytes, err := httputil.DumpResponse(resp, true)
			if err == nil {
				t.Cache.Set(cacheKey, respBytes)
			}
		}
	} else {
		if t.Debug {
			log.WithFields(log.Fields{
				"cacheable":   cacheable,
				"cacheKey":    cacheKey,
				"resp.Header": resp.Header,
			}).Warn("httpcache.RoundTrip() Delete")
		}
		t.Cache.Delete(cacheKey)
	}
	return resp, nil
}

type ETagTransport struct {
	ETag string
}

func (t *ETagTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.ETag == "" {
		return nil, errors.New("t.ETag is empty")
	}

	// To set extra querystring params, we must make a copy of the Request so
	// that we don't modify the Request we were given. This is required by the
	// specification of http.RoundTripper.
	req = cloneRequest(req)
	req.Header.Add("if-none-match", t.ETag)

	// Make the HTTP request.
	resp, err := http.DefaultTransport.RoundTrip(req)
	log.Println("etag: ", resp.Header.Get("etag"))
	return resp, err
}
