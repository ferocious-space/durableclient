package durableclient

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/ferocious-space/durableclient/cleanhttp"

	"github.com/ferocious-space/durableclient/middlewares"
)

func NewDurableClient(logger logr.Logger, agent string, opt ...ClientOptions) *DurableClient {
	dc := &DurableClient{ctx: context.TODO(), logger: logger, agent: agent, http2: true}
	for _, o := range opt {
		o(dc)
	}
	client := new(http.Client)
	if dc.pooled {
		if dc.http2 {
			client = cleanhttp.DefaultPooledClient()
		} else {
			client = cleanhttp.DefaultPooledClient(cleanhttp.WithHTTP2Disabled())
		}
	} else {
		if dc.http2 {
			client = cleanhttp.DefaultClient()
		} else {
			client = cleanhttp.DefaultClient(cleanhttp.WithHTTP2Disabled())
		}
	}
	dc.ctx = logr.NewContext(dc.ctx, dc.logger)
	if dc.cache == nil {
		dc.Client = middlewares.Enable(
			dc.retrier, middlewares.Drainer(dc.ctx),
		).ThenMiddleware(middlewares.Agent(dc.agent)).
			ExtendWith(
				middlewares.Enable(
					dc.retrier, middlewares.Retrier(5),
				),
				middlewares.Enable(
					dc.retrier, middlewares.NewCircuitMiddleware().Middleware(80, time.Second*60),
				),
			).ThenClient(client, true)
	} else {
		dc.Client = middlewares.Cache(dc.ctx, dc.cache).
			ExtendWith(
				middlewares.Agent(dc.agent),
				middlewares.Enable(
					dc.retrier, middlewares.Retrier(5),
				),
				middlewares.Enable(
					dc.retrier, middlewares.NewCircuitMiddleware().Middleware(80, time.Second*60),
				),
			).ThenClient(client, true)
	}
	return dc
}

func (c *DurableClient) Download(url string, w io.Writer) (n int64, err error) {
	rsp, err := c.Get(url)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = rsp.Body.Close()
	}()

	n, err = io.Copy(w, rsp.Body)
	if err != nil {
		return 0, err
	}
	if n != rsp.ContentLength {
		return n, errors.Errorf("contentLength %d dont match written bytes %s", rsp.ContentLength, n)
	}
	return n, nil
}
