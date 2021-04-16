package durableclient

import (
	"context"
	"net/http"
	"time"

	"github.com/ferocious-space/durableclient/cleanhttp"
	"github.com/go-logr/logr"

	"github.com/ferocious-space/durableclient/middlewares"
)

func NewDurableClient(ctx context.Context, logger logr.Logger, agent string) *DurableClient {
	return &DurableClient{ctx: ctx, logger: logger, agent: agent, http2: true}
}

func (c *DurableClient) Client(opt ...ClientOptions) *http.Client {
	for _, o := range opt {
		o(c)
	}
	client := new(http.Client)
	if c.pooled {
		if c.http2 {
			client = cleanhttp.DefaultPooledClient()
		} else {
			client = cleanhttp.DefaultPooledClient(cleanhttp.WithHTTP2Disabled())
		}
	} else {
		if c.http2 {
			client = cleanhttp.DefaultClient()
		} else {
			client = cleanhttp.DefaultClient(cleanhttp.WithHTTP2Disabled())
		}
	}
	c.ctx = logr.NewContext(c.ctx, c.logger)
	if c.cache == nil {
		return middlewares.Drainer(c.ctx).
			ThenMiddleware(middlewares.Agent(c.agent)).
			ExtendWith(
				middlewares.Enable(c.retrier, middlewares.Retrier(5)),
				middlewares.NewCircuitMiddleware().Middleware(80, time.Second*60),
			).ThenClient(client, true)
	}
	return middlewares.Cache(c.ctx, c.cache).
		ExtendWith(
			middlewares.Agent(c.agent),
			middlewares.Enable(c.retrier, middlewares.Retrier(5)),
			middlewares.NewCircuitMiddleware().Middleware(80, time.Second*60),
		).ThenClient(client, true)
}
