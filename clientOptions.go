package durableclient

import (
	"context"

	"github.com/ferocious-space/httpcache"
)

type ClientOptions func(c *DurableClient)

// OptionCache attaches HTTPCache to the client
func OptionCache(cache httpcache.Cache) ClientOptions {
	return func(c *DurableClient) {
		c.setCache(cache)
	}
}

func OptionConnectionPooling() ClientOptions {
	return func(c *DurableClient) {
		c.setPooled(true)
	}
}

func OptionContext(ctx context.Context) ClientOptions {
	return func(c *DurableClient) {
		c.setCTX(ctx)
	}
}

func OptionRetrier() ClientOptions {
	return func(c *DurableClient) {
		c.setRetrier(true)
	}
}

func OptionHTTP2Disabled() ClientOptions {
	return func(c *DurableClient) {
		c.setHTTP2(false)
	}
}
