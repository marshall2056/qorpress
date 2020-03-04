package httpcache

import (
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type CustomTransport struct {
	Transport http.RoundTripper
	MaxStale  int //seconds
	Debug     bool
}

func (c *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := c.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	req2 := cloneRequest(req) // per RoundTripper contract
	req2.Header.Add("Cache-Control", "Max-Stale="+strconv.Itoa(c.MaxStale))
	if c.Debug {
		log.WithFields(log.Fields{
			"Max-Stale": c.MaxStale,
		}).Warn("httpcache.CustomTransport.RoundTrip()")
	}
	res, err := transport.RoundTrip(req2)
	return res, err
}
