package httpcache

import (
	"bufio"
	"bytes"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
)

type pending struct {
	req *http.Request
	buf *bytes.Buffer
	err error
	wg  sync.WaitGroup
}

type BlockingTransport struct {
	mu        sync.RWMutex
	pending   map[string]*pending
	Transport http.RoundTripper
	debug     bool
}

func NewBlockingTransport(rt http.RoundTripper) http.RoundTripper {
	t := BlockingTransport{
		pending:   make(map[string]*pending),
		Transport: rt,
	}
	return &t
}

func (t *BlockingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	key := cacheKey(req)

	if t.debug {
		log.WithFields(log.Fields{
			"key": key,
		}).Warn("httpcache.BlockingTransport.RoundTrip()")
	}

	if key == "" {
		return t.transport().RoundTrip(req)
	}
	var p *pending
	t.mu.RLock()
	p = t.pending[key]
	t.mu.RUnlock()
	if p != nil {
		return p.Response()
	}
	t.mu.Lock()
	if p = t.pending[key]; p != nil {
		t.mu.Unlock()
		return p.Response()
	}
	p = &pending{
		req: req,
		buf: new(bytes.Buffer),
	}
	p.wg.Add(1)
	t.pending[key] = p
	t.mu.Unlock()

	go func() {
		var resp *http.Response
		if resp, p.err = t.transport().RoundTrip(p.req); p.err == nil {
			p.err = resp.Write(p.buf)
		}
		t.mu.Lock()
		delete(t.pending, key)
		t.mu.Unlock()
		p.wg.Done()
	}()
	return p.Response()
}

func (p *pending) Response() (*http.Response, error) {
	p.wg.Wait()
	if p.err != nil {
		return nil, p.err
	}
	data := p.buf.Bytes()
	r := bufio.NewReaderSize(bytes.NewReader(data), len(data))
	return http.ReadResponse(r, p.req)
}
