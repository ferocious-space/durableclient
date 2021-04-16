package durableclient

import (
	"context"

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
	ctx     context.Context
	logger  logr.Logger
	agent   string
	cache   httpcache.Cache
	pooled  bool
	retrier bool
	http2   bool
}

func (c *DurableClient) SetHttp2(http2 bool) {
	c.http2 = http2
}

func (c *DurableClient) SetRetrier(retrier bool) {
	c.retrier = retrier
}

func (c *DurableClient) SetCtx(ctx context.Context) {
	c.ctx = ctx
}

func (c *DurableClient) SetCache(cache httpcache.Cache) {
	c.cache = cache
}

func (c *DurableClient) SetPooled(pooled bool) {
	c.pooled = pooled
}

func (c *DurableClient) Clone() *DurableClient {
	cloned := cloneClient{c}
	return cloned.Clone()
}
