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
	return &DurableClient{ctx: ctx, logger: logger, agent: agent}
}

func (c *DurableClient) Client() *http.Client {
	client := new(http.Client)
	if c.pooled {
		client = cleanhttp.DefaultPooledClient()
	} else {
		client = cleanhttp.DefaultClient()
	}
	c.ctx = logr.NewContext(c.ctx, c.logger)
	if c.cache == nil {
		return middlewares.Drainer(c.ctx).ThenMiddleware(middlewares.Agent(c.agent)).ExtendWith(middlewares.NewCircuitMiddleware().Middleware(80, time.Second*60)).ThenClient(client, true)
	}
	return middlewares.Cache(c.ctx, c.cache).ExtendWith(middlewares.Agent(c.agent), middlewares.NewCircuitMiddleware().Middleware(80, time.Second*60)).ThenClient(client, true)
}
