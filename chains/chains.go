package chains

import (
	"net/http"
	"net/http/cookiejar"
	"sync"

	"github.com/ferocious-space/durableclient/cleanhttp"
	"golang.org/x/net/publicsuffix"
)

var httpClient *http.Client
var cmu sync.Mutex

var _ = defaultClient()

func defaultClient() *http.Client {
	cmu.Lock()
	defer cmu.Unlock()
	if httpClient == nil {
		jar, err := cookiejar.New(&cookiejar.Options{
			PublicSuffixList: publicsuffix.List,
		})
		if err != nil {
			panic(err.Error())
		}

		client := cleanhttp.DefaultClient()
		client.Jar = jar
		httpClient = client
	}
	return httpClient
}

// RoundTripFunc wraps a func to make it into an http.RoundTripper. Similar to http.HandleFunc.
type RoundTripFunc func(*http.Request) (*http.Response, error)

// RoundTrip implements RoundTripper interface
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// Middleware represents an http client-side middleware.
type Middleware func(http.RoundTripper) http.RoundTripper

func (r Middleware) RoundTrip(req *http.Request) (*http.Response, error) {
	return r(defaultClient().Transport).RoundTrip(req)
}

func (r Middleware) ThenClient(client *http.Client, clone bool) *http.Client {
	if client == nil {
		client = defaultClient()
	}
	if clone {
		c := *client
		client = &c
	}
	if client.Transport == nil {
		client.Transport = defaultClient().Transport
	}
	client.Transport = r(client.Transport)
	return client
}

func (r Middleware) DefaultClient() *http.Client {
	return r.ThenClient(nil, false)
}

func (r Middleware) ThenRoundTripper(t http.RoundTripper) http.RoundTripper {
	if t == nil {
		t = defaultClient().Transport
	}
	return r(t)
}

func (r Middleware) DefaultRoundTripper() http.RoundTripper {
	return r.ThenRoundTripper(nil)
}

func (r Middleware) ThenMiddleware(m Middleware) Chain {
	return NewChain(r, m)
}

type Chain struct {
	middlewares []Middleware
}

func NewChain(middlewares ...Middleware) Chain {
	return Chain{append([]Middleware(nil), middlewares...)}
}

func (c Chain) RoundTrip(req *http.Request) (*http.Response, error) {
	return c.ThenRoundTripper(defaultClient().Transport).RoundTrip(req)
}

func (c Chain) ThenRoundTripper(t http.RoundTripper) http.RoundTripper {
	if t == nil {
		t = defaultClient().Transport
	}
	for i := range c.middlewares {
		t = c.middlewares[len(c.middlewares)-1-i](t)
	}
	return t
}

func (c Chain) DefaultRoundTripper() http.RoundTripper {
	return c.ThenRoundTripper(nil)
}

func (c Chain) ThenClient(client *http.Client, clone bool) *http.Client {
	if client == nil {
		client = defaultClient()
	}
	if clone {
		c := *client
		client = &c
	}
	client.Transport = c.ThenRoundTripper(client.Transport)
	return client
}

func (c Chain) DefaultClient(clone bool) *http.Client {
	return c.ThenClient(nil, clone)
}

func (c Chain) ExtendWith(middlewares ...Middleware) Chain {
	newChain := make([]Middleware, 0, len(c.middlewares)+len(middlewares))
	newChain = append(newChain, c.middlewares...)
	newChain = append(newChain, middlewares...)
	return Chain{newChain}
}
