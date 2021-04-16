package durableclient

import (
	"context"

	"github.com/ferocious-space/httpcache"
)

type ClientOptions func(c *DurableClient)

// WithCache attaches HTTPCache to the client
func WithCache(cache httpcache.Cache) ClientOptions {
	return func(c *DurableClient) {
		c.SetCache(cache)
	}
}

func WithConnectionPooling(pooling bool) ClientOptions {
	return func(c *DurableClient) {
		c.SetPooled(pooling)
	}
}

func WithContext(ctx context.Context) ClientOptions {
	return func(c *DurableClient) {
		c.SetCtx(ctx)
	}
}

func WithRetrier() ClientOptions {
	return func(c *DurableClient) {
		c.SetRetrier(true)
	}
}
