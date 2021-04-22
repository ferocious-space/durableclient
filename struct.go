package durableclient

import (
	"context"
	"net/http"

	"github.com/ferocious-space/httpcache"
	"github.com/go-logr/logr"
)

type cloneClient struct {
	*DurableClient
}

func (x *cloneClient) Clone() *DurableClient {
	c := *x.DurableClient
	return &c
}

type DurableClient struct {
	*http.Client

	ctx     context.Context
	logger  logr.Logger
	agent   string
	cache   httpcache.Cache
	pooled  bool
	retrier bool
	http2   bool
}

func (c *DurableClient) setHTTP2(http2 bool) {
	c.http2 = http2
}

func (c *DurableClient) setRetrier(retrier bool) {
	c.retrier = retrier
}

func (c *DurableClient) setCTX(ctx context.Context) {
	c.ctx = ctx
}

func (c *DurableClient) setCache(cache httpcache.Cache) {
	c.cache = cache
}

func (c *DurableClient) setPooled(pooled bool) {
	c.pooled = pooled
}

func (c *DurableClient) Clone(opt ...ClientOptions) *DurableClient {
	cloned := cloneClient{c}
	client := cloned.Clone()
	for _, o := range opt {
		o(client)
	}
	return client
}

func (c *DurableClient) HttpClient() *http.Client {
	return &http.Client{
		Transport: c.Transport,
	}
}

// func (c *DurableClient) httpDo(ctx context.Context, req *http.Request, f func(resp *http.Response, err error) error) error {
// 	ch := make(chan error, 1)
// 	req = req.WithContext(ctx)
// 	go func() { ch <- f(c.Do(req)) }()
// 	select {
// 	case <-ctx.Done():
// 		<-ch
// 		return ctx.Err()
// 	case err := <-ch:
// 		return err
// 	}
// }
//
// func (c *DurableClient) GoGet(ctx context.Context, url string) (<-chan *http.Response, <-chan error) {
// 	respChan := make(chan *http.Response)
// 	errChan := make(chan error)
// 	go func() {
// 		defer close(respChan)
// 		defer close(errChan)
// 		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
// 		if err != nil {
// 			respChan <- nil
// 			errChan <- err
// 			return
// 		}
// 		errChan <- c.httpDo(ctx, req, func(resp *http.Response, err error) error {
// 			respChan <- resp
// 			return err
// 		})
//
// 	}()
// 	return respChan, errChan
// }
