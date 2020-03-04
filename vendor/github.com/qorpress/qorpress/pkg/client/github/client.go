/*
 * MIT License
 *
 * Copyright (c) 2017 SmartestEE Co., Ltd..
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

/*
 * Revision History:
 *     Initial: 2017/08/08        Jia Chenhui
 */

// refs.
// - https://github.com/fengyfei/nuts/blob/master/github/client/client.go
// - https://github.com/michalskiluc/Andariel/blob/master/cmd/github/process/repo.go
// - https://github.com/michalskiluc/Andariel/blob/master/pkg/github/repo.go

package client

import (
	"context"
	"errors"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/go-github/v29/github"
	"github.com/k0kubun/pp"
	"github.com/sirupsen/logrus"
	httpcache "github.com/sniperkit/cacher"
	"github.com/x0rzkov/httpcache/backend/diskcache"
	"golang.org/x/oauth2"
)

type rateLimitCategory uint8

var log *logrus.Logger

func init() {
	log = &logrus.Logger{
		Out: os.Stderr,
		Formatter: &logrus.TextFormatter{
			DisableTimestamp: true,
		},
	}
}

const (
	coreCategory rateLimitCategory = iota
	searchCategory
	categories
)

type GHClient struct {
	CachePath  string
	Client     *github.Client
	Manager    *ClientManager
	rateLimits [categories]Rate
	timer      *time.Timer
	rateMu     sync.Mutex
}

type Rate struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

// NewManager create a new client manager based on tokens.
func NewManager(cachePath string, tokens []string) *ClientManager {
	var cm *ClientManager = &ClientManager{
		reclaim:  make(chan *GHClient),
		Dispatch: make(chan *GHClient, len(tokens)),
		shutdown: make(chan struct{}),
	}

	clients := newClients(cachePath, tokens)

	go cm.start()
	go func() {
		for _, c := range clients {
			// pp.Println("c, ", c)
			pp.Println("c.isLimited():: ", c.isLimited())
			if !c.isLimited() {
				c.Manager = cm
				cm.reclaim <- c
			}
		}
	}()

	pp.Println(cm)
	return cm
}

// newClient create client based on token.
func newClient(cachePath, token string) (client *GHClient, err error) {
	if token == "" {
		client = new(GHClient)
		tokenSource := new(oauth2.TokenSource)
		if !client.init(cachePath, *tokenSource) {
			err = errors.New("failed to create client")
			log.Fatalln("client.tokenSource: ", *tokenSource)
			return nil, err
		}

		return client, nil
	}

	client = new(GHClient)
	// client.CachePath = cachePath
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	if !client.init(cachePath, tokenSource) {
		err = errors.New("failed to create client")
		log.Fatalln("client.tokenSource: ", tokenSource)
		return nil, err
	}
	return client, nil
}

// init initializes the client, returns true if available, or returns false.
func (c *GHClient) init(cachePath string, tokenSource oauth2.TokenSource) bool {

	// httpClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	// ghClient := github.NewClient(httpClient)

	cache := diskcache.New(cachePath)
	cacheTransport := httpcache.NewTransport(cache)
	authTransport := oauth2.Transport{Source: tokenSource, Base: cacheTransport}
	httpClient := &http.Client{Transport: &authTransport}
	ghClient := github.NewClient(httpClient)

	c.CachePath = cachePath
	c.Client = ghClient

	// pp.Println(ghClient)

	if !c.isValidToken(httpClient) {
		log.Fatalln("isInValidToken: ", c.isValidToken(httpClient))
		return false
	}

	if c.isLimited() {
		log.Fatalln("isLimited: ", c.isLimited())
		return false
	}

	return true
}

// isValidToken check if token is valid.
func (c *GHClient) isValidToken(httpClient *http.Client) bool {
	resp, err := c.makeRequest(httpClient)
	if err != nil {
		return false
	}

	err = github.CheckResponse(resp)
	if _, ok := err.(*github.TwoFactorAuthError); ok {
		return false
	}

	return true
}

// makeRequest sends an HTTP GET request and returns an HTTP response, following
// policy (such as redirects, cookies, auth) as configured on the client.
func (c *GHClient) makeRequest(httpClient *http.Client) (*http.Response, error) {
	req, err := c.Client.NewRequest("GET", "", nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// isLimited check if the client is available.
func (c *GHClient) isLimited() bool {
	rate, _, err := c.Client.RateLimits(context.Background())
	if err != nil {
		return true
	}

	response := new(struct {
		Resource *github.RateLimits `json:"resource"`
	})
	response.Resource = rate

	if response.Resource != nil {
		c.rateMu.Lock()
		defer c.rateMu.Unlock()
		if response.Resource.Core != nil {
			c.rateLimits[coreCategory].Limit = response.Resource.Core.Limit
			c.rateLimits[coreCategory].Remaining = response.Resource.Core.Remaining
			c.rateLimits[coreCategory].Reset = response.Resource.Core.Reset.Time
			return false
		}
		if response.Resource.Search != nil {
			c.rateLimits[searchCategory].Remaining = response.Resource.Search.Remaining
			c.rateLimits[searchCategory].Limit = response.Resource.Search.Limit
			c.rateLimits[searchCategory].Reset = response.Resource.Search.Reset.Time
			return false
		}
	}

	return true
}

// initTimer initialize client timer.
func (c *GHClient) initTimer(resp *github.Response) {
	if resp != nil {
		timer := time.NewTimer((*resp).Rate.Reset.Time.Sub(time.Now()) + time.Second*2)
		c.timer = timer

		return
	}
}

// newClients create a client list based on tokens.
func newClients(cachePath string, tokens []string) []*GHClient {
	var clients []*GHClient
	for _, t := range tokens {
		pp.Println("cachePath:", cachePath, "token:", t)
		client, err := newClient(cachePath, t)
		if err != nil {
			log.Warnln("error while creating new client: ", err)
			continue
		}
		clients = append(clients, client)
	}
	pp.Println("clients.length", len(clients))
	return clients
}

// ClientManager used to manage the valid client.
type ClientManager struct {
	Dispatch chan *GHClient
	reclaim  chan *GHClient
	shutdown chan struct{}
}

// start start reclaim and dispatch the client.
func (cm *ClientManager) start() {
	for {
		select {
		case v := <-cm.reclaim:
			cm.Dispatch <- v
		case <-cm.shutdown:
			close(cm.Dispatch)
			close(cm.reclaim)
			return
		}
	}
}

// Fetch fetch a valid client.
func (cm *ClientManager) Fetch() *GHClient {
	log.Warnln("Fetch...", cm.Dispatch)
	return <-cm.Dispatch
}

// Reclaim reclaim client while the client is valid.
// resp: The response returned when calling the client.
func Reclaim(client *GHClient, resp *github.Response) {
	client.initTimer(resp)
	select {
	case <-client.timer.C:
		client.Manager.reclaim <- client
	}
}

// Shutdown shutdown the client manager.
func (cm *ClientManager) Shutdown() {
	close(cm.shutdown)
}
