package httpcache

import (
	"errors"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// ErrNoDateHeader indicates that the HTTP headers contained no Date header.
var (
	clock           timer = &realClock{}
	ErrNoDateHeader       = errors.New("No Date header")
)

// headerAllCommaSepValues returns all comma-separated values (each
// with whitespace trimmed) for header name in headers. According to
// Section 4.2 of the HTTP/1.1 spec
// (http://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2),
// values from multiple occurrences of a header should be concatenated, if
// the header's value is a comma-separated list.
func headerAllCommaSepValues(headers http.Header, name string) []string {
	var vals []string
	for _, val := range headers[http.CanonicalHeaderKey(name)] {
		fields := strings.Split(val, ",")
		for i, f := range fields {
			fields[i] = strings.TrimSpace(f)
		}
		vals = append(vals, fields...)
	}
	if Debug {
		log.WithFields(log.Fields{
			"name": name,
			"vals": vals,
		}).Info("httpcache.headerAllCommaSepValues()")
	}
	return vals
}

// varyMatches will return false unless all of the cached values for the headers listed in Vary
// match the new request
func varyMatches(cachedResp *http.Response, req *http.Request) bool {
	for _, header := range headerAllCommaSepValues(cachedResp.Header, "Vary") {
		header = http.CanonicalHeaderKey(header)
		if Debug {
			log.WithFields(log.Fields{
				"header": header,
				"cachedResp.Header.Get(X-Varied-" + header + ")": cachedResp.Header.Get("X-Varied-" + header),
				"req.Header.Get(header)":                         req.Header.Get(header),
				"vary": (header != "" && req.Header.Get(header) != cachedResp.Header.Get("X-Varied-"+header)),
			}).Info("httpcache.varyMatches(), cachedResp")
		}
		if header != "" && req.Header.Get(header) != cachedResp.Header.Get("X-Varied-"+header) {
			return false
		}
	}
	return true
}

// Date parses and returns the value of the Date header.
func Date(respHeaders http.Header) (date time.Time, err error) {
	dateHeader := respHeaders.Get("Date")
	if dateHeader == "" {
		err = ErrNoDateHeader
		return
	}
	if Debug {
		log.WithFields(log.Fields{
			"dateHeader": dateHeader,
		}).Info("httpcache.Date(), RESP")
	}
	return time.Parse(time.RFC1123, dateHeader)
}

type realClock struct{}

func (c *realClock) since(d time.Time) time.Duration {
	return time.Since(d)
}

type timer interface {
	since(d time.Time) time.Duration
}

// getFreshness will return one of fresh/stale/transparent based on the cache-control
// values of the request and the response
//
// fresh indicates the response can be returned
// stale indicates that the response needs validating before it is returned
// transparent indicates the response should not be used to fulfil the request
//
// Because this is only a private cache, 'public' and 'private' in cache-control aren't
// signficant. Similarly, smax-age isn't used.
func getFreshness(respHeaders, reqHeaders http.Header) (freshness int) {
	respCacheControl := parseCacheControl(respHeaders)
	reqCacheControl := parseCacheControl(reqHeaders)
	if _, ok := reqCacheControl["No-Cache"]; ok {
		if Debug {
			log.WithFields(log.Fields{
				"No-Cache":    ok,
				"transparent": transparent,
			}).Info("httpcache.getFreshness(), REQ")
		}
		return transparent
	}
	if _, ok := respCacheControl["No-Cache"]; ok {
		if Debug {
			log.WithFields(log.Fields{
				"No-Cache": ok,
				"state":    stale,
			}).Info("httpcache.getFreshness(), RESP")
		}
		return stale
	}
	if _, ok := reqCacheControl["Only-If-Cached"]; ok {
		if Debug {
			log.WithFields(log.Fields{
				"Only-If-Cached": ok,
				"fresh":          fresh,
			}).Info("httpcache.getFreshness(), REQ")
		}
		return fresh
	}

	date, err := Date(respHeaders)
	if err != nil {
		return stale
	}
	currentAge := clock.since(date)

	var lifetime time.Duration
	var zeroDuration time.Duration

	// If a response includes both an Expires header and a max-age directive,
	// the max-age directive overrides the Expires header, even if the Expires header is more restrictive.
	if maxAge, ok := respCacheControl["Max-Age"]; ok {
		lifetime, err = time.ParseDuration(maxAge + "s")
		if err != nil {
			lifetime = zeroDuration
		}
	} else {
		expiresHeader := respHeaders.Get("Expires")
		if expiresHeader != "" {
			expires, err := time.Parse(time.RFC1123, expiresHeader)
			if err != nil {
				lifetime = zeroDuration
			} else {
				lifetime = expires.Sub(date)
			}
		}
	}

	if maxAge, ok := reqCacheControl["Max-Age"]; ok {
		// the client is willing to accept a response whose age is no greater than the specified time in seconds
		lifetime, err = time.ParseDuration(maxAge + "s")
		if err != nil {
			lifetime = zeroDuration
		}
	}

	if minfresh, ok := reqCacheControl["Min-Fresh"]; ok {
		//  the client wants a response that will still be fresh for at least the specified number of seconds.
		minfreshDuration, err := time.ParseDuration(minfresh + "s")
		if err == nil {
			currentAge = time.Duration(currentAge + minfreshDuration)
		}
	}

	if maxstale, ok := reqCacheControl["Max-Stale"]; ok {
		// Indicates that the client is willing to accept a response that has exceeded its expiration time.
		// If max-stale is assigned a value, then the client is willing to accept a response that has exceeded
		// its expiration time by no more than the specified number of seconds.
		// If no value is assigned to max-stale, then the client is willing to accept a stale response of any age.
		//
		// Responses served only because of a max-stale value are supposed to have a Warning header added to them,
		// but that seems like a  hassle, and is it actually useful? If so, then there needs to be a different
		// return-value available here.
		if maxstale == "" {
			return fresh
		}
		maxstaleDuration, err := time.ParseDuration(maxstale + "s")
		if err == nil {
			currentAge = time.Duration(currentAge - maxstaleDuration)
		}
	}

	if lifetime > currentAge {
		return fresh
	}

	return stale
}

// Returns true if either the request or the response includes the stale-if-error
// cache control extension: https://tools.ietf.org/html/rfc5861
func canStaleOnError(respHeaders, reqHeaders http.Header) bool {
	respCacheControl := parseCacheControl(respHeaders)
	reqCacheControl := parseCacheControl(reqHeaders)

	var err error
	lifetime := time.Duration(-1)

	if staleMaxAge, ok := respCacheControl["Stale-If-Error"]; ok {
		if staleMaxAge != "" {
			lifetime, err = time.ParseDuration(staleMaxAge + "s")
			if err != nil {
				return false
			}
		} else {
			return true
		}
	}

	if staleMaxAge, ok := reqCacheControl["Stale-If-Error"]; ok {
		if staleMaxAge != "" {
			lifetime, err = time.ParseDuration(staleMaxAge + "s")
			if err != nil {
				return false
			}
		} else {
			return true
		}
	}

	if lifetime >= 0 {
		date, err := Date(respHeaders)
		if err != nil {
			return false
		}
		currentAge := clock.since(date)
		if lifetime > currentAge {
			return true
		}
	}

	return false
}

func getEndToEndHeaders(respHeaders http.Header) []string {
	// These headers are always hop-by-hop
	hopByHopHeaders := map[string]struct{}{
		"Connection":          struct{}{},
		"Keep-Alive":          struct{}{},
		"Proxy-Authenticate":  struct{}{},
		"Proxy-Authorization": struct{}{},
		"Te":                struct{}{},
		"Trailers":          struct{}{},
		"Transfer-Encoding": struct{}{},
		"Upgrade":           struct{}{},
	}

	for _, extra := range strings.Split(respHeaders.Get("Connection"), ",") {
		// any header listed in connection, if present, is also considered hop-by-hop
		if strings.Trim(extra, " ") != "" {
			hopByHopHeaders[http.CanonicalHeaderKey(extra)] = struct{}{}
		}
	}
	endToEndHeaders := []string{}
	for respHeader, _ := range respHeaders {
		if _, ok := hopByHopHeaders[respHeader]; !ok {
			endToEndHeaders = append(endToEndHeaders, respHeader)
		}
	}
	return endToEndHeaders
}

// func canStore(reqCacheControl, respCacheControl cacheControl) (canStore bool) {
func canStore(code int, reqCacheControl, respCacheControl cacheControl) (canStore bool) {
	if _, ok := cacheableResponseCodes[code]; !ok {
		return false
	}
	if _, ok := respCacheControl["no-store"]; ok {
		return false
	}
	if _, ok := reqCacheControl["no-store"]; ok {
		return false
	}
	return true
}
