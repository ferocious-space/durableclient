package durableclient

import (
	"context"

	"github.com/ferocious-space/httpcache"
	"github.com/go-logr/logr"
)

type DurableClient struct {
	ctx    context.Context
	logger logr.Logger
	agent  string
	cache  httpcache.Cache
	pooled bool
}

func (c *DurableClient) SetCache(cache httpcache.Cache) {
	c.cache = cache
}

func (c *DurableClient) SetPooled(pooled bool) {
	c.pooled = pooled
}

func (c *DurableClient) WithCache(cache httpcache.Cache) *DurableClient {
	c.SetCache(cache)
	return c
}

func (c *DurableClient) WithPool(pool bool) *DurableClient {
	c.SetPooled(pool)
	return c
}
