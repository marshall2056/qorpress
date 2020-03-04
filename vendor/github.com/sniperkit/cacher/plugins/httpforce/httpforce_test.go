package httpforcecache

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sniperkit/httpcache"
	"github.com/stretchr/testify/assert"
)

var s struct {
	server    *httptest.Server
	client    http.Client
	transport *Transport
	count     int
}

func TestMain(m *testing.M) {
	flag.Parse()
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	tp := NewMemoryCacheTransport()
	client := http.Client{Transport: tp}
	s.transport = tp
	s.client = client
	s.count = 0

	mux := http.NewServeMux()
	s.server = httptest.NewServer(mux)

	mux.HandleFunc("/no-cache", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.count++
		fmt.Fprintf(w, "%d", s.count)
	}))
	mux.HandleFunc("/not-found", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.count++
		msg := fmt.Sprintf("404 : %d", s.count)
		http.Error(w, msg, http.StatusNotFound)
	}))
}

func teardown() {
	s.server.Close()
}

func resetTest() {
	s.transport.Cache = httpcache.NewMemoryCache()
	s.count = 0
}

func getResponseText(t *testing.T, resp *http.Response) string {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	err = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	text := buf.String()
	return text
}

func request(t *testing.T, url string) string {
	req, err := http.NewRequest("GET", s.server.URL+url, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	text := getResponseText(t, resp)
	return text
}

func TestSingleRequest(t *testing.T) {
	resetTest()
	{
		text := request(t, "/no-cache")
		assert.Equal(t, "1", text)
	}
}

func TestMultipleRequest(t *testing.T) {
	resetTest()
	{
		for i := 0; i < 2; i++ {
			text := request(t, "/no-cache")
			assert.Equal(t, "1", text)
		}
	}
}

func TestDeleteCache(t *testing.T) {
	resetTest()
	{
		text := request(t, "/no-cache")
		assert.Equal(t, "1", text)

		req, _ := http.NewRequest("GET", s.server.URL+"/no-cache", nil)
		s.transport.DeleteCache(req)

		t2 := request(t, "/no-cache")
		assert.Equal(t, "2", t2)
	}
}

// do not caching error response
func TestErrorPage(t *testing.T) {
	resetTest()
	{
		t1 := request(t, "/not-found")
		assert.Equal(t, "404 : 1\n", t1)

		t2 := request(t, "/not-found")
		assert.Equal(t, "404 : 2\n", t2)
	}
}
