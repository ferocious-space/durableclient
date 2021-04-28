package durableclient

import (
	"context"
	"net/http"
	"time"

	"github.com/go-logr/logr"

	"github.com/ferocious-space/durableclient/chains"
	"github.com/ferocious-space/durableclient/cleanhttp"

	"github.com/ferocious-space/durableclient/middlewares"
)

//NewDurableClient creates default opinionated client
//
//params := &durableOption{
//		cache:         nil,
//		ctx:           context.TODO(),
//		logger:        logr.Discard(),
//		agent:         "https://github.com/ferocious-space/durableclient",
//		pooling:       false,
//		retrier:       true,
//		numRetries:    3,
//		circuit:       true,
//		maxErrors:     80,
//      ErrorThresholdPercentage: 100,
//		rollingWindow: time.Second * 60,
//	}
// use ClientOptions to change them
func NewDurableClient(opt ...ClientOptions) *http.Client {

	params := &durableOption{
		cache:                    nil,
		ctx:                      context.TODO(),
		logger:                   logr.Discard(),
		agent:                    "https://github.com/ferocious-space/durableclient",
		pooling:                  false,
		retrier:                  true,
		numRetries:               3,
		circuit:                  true,
		maxErrors:                80,
		ErrorThresholdPercentage: 100,
		rollingWindow:            time.Second * 60,
		opt:                      []cleanhttp.TransportOptions{},
	}

	for _, o := range opt {
		o(params)
	}
	client := cleanhttp.DefaultClient(params.opt...)

	// start the builder with drainer if we have pooling ( its important to drain connections to be able to reclaim them )
	builder := chains.NewChain(params.ctx, middlewares.Enable(params.pooling, middlewares.Drainer()))

	// add the cache if not nil
	withCache := params.cache != nil
	builder = builder.ExtendWith(middlewares.Enable(withCache, middlewares.Cache(params.cache)))

	// add the agent
	builder = builder.ExtendWith(middlewares.Agent(params.agent))

	// add retrier if needed
	builder = builder.ExtendWith(middlewares.Enable(params.retrier, middlewares.Retrier(params.numRetries)))

	// add circuit if defined
	builder = builder.ExtendWith(
		middlewares.Enable(
			params.circuit,
			middlewares.NewCircuitMiddleware().Middleware(params.maxErrors, params.ErrorThresholdPercentage, params.rollingWindow),
		),
	)

	return builder.ThenClient(client, false)
}
