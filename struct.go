package durableclient

import (
	"context"

	"github.com/ferocious-space/httpcache"
	"github.com/go-logr/logr"
)

type DurableClient struct {
	ctx     context.Context
	logger  logr.Logger
	agent   string
	cache   httpcache.Cache
	pooled  bool
	retrier bool
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
	cloned := CloneClient{c}
	return cloned.Clone()
}

func (c *DurableClient) WithCache(cache httpcache.Cache) *DurableClient {
	c.SetCache(cache)
	return c
}

func (c *DurableClient) WithPool(pool bool) *DurableClient {
	c.SetPooled(pool)
	return c
}

type CloneClient struct {
	*DurableClient
}

func (x *CloneClient) Clone() *DurableClient {
	c := *x.DurableClient
	return &c
}
