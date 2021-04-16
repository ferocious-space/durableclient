package durableclient

import (
	"context"

	"github.com/ferocious-space/httpcache"
)

type ClientOptions func(c *DurableClient)

// OptionCache attaches HTTPCache to the client
func OptionCache(cache httpcache.Cache) ClientOptions {
	return func(c *DurableClient) {
		c.SetCache(cache)
	}
}

func OptionConnectionPooling() ClientOptions {
	return func(c *DurableClient) {
		c.SetPooled(true)
	}
}

func OptionContext(ctx context.Context) ClientOptions {
	return func(c *DurableClient) {
		c.SetCtx(ctx)
	}
}

func OptionRetrier() ClientOptions {
	return func(c *DurableClient) {
		c.SetRetrier(true)
	}
}

func OptionHTTP2Disabled() ClientOptions {
	return func(c *DurableClient) {
		c.SetHttp2(false)
	}
}
