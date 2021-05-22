package durableclient

import (
	"net/http"
	"time"

	"github.com/cep21/circuit/v3/closers/hystrix"
	"github.com/cep21/circuit/v3/metrics/rolling"
	"github.com/go-logr/logr"

	"github.com/ferocious-space/durableclient/chains"
	"github.com/ferocious-space/durableclient/cleanhttp"

	"github.com/ferocious-space/durableclient/middlewares"
)

//NewDurableClient creates default opinionated client
//
//params := &durableOption{
//		cache:         nil,
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
		cache:                 nil,
		logger:                logr.Discard(),
		agent:                 "https://github.com/ferocious-space/durableclient",
		pooling:               false,
		retrier:               true,
		numRetries:            3,
		circuit:               true,
		stream:                false,
		maxConcurrentRequests: 50,

		closerConfig: &hystrix.ConfigureCloser{
			SleepWindow:                  time.Duration(5) * time.Minute,
			HalfOpenAttempts:             int64((time.Duration(5)*time.Minute)/time.Second) / 5,
			RequiredConcurrentSuccessful: 3,
		},
		openerConfig: &hystrix.ConfigureOpener{
			ErrorThresholdPercentage: 50,
			RequestVolumeThreshold:   80,
			RollingDuration:          time.Duration(1) * time.Minute,
			NumBuckets:               int(int64(time.Minute/time.Millisecond) / 100),
		},
		statsConfig: &rolling.RunStatsConfig{
			RollingStatsDuration:        time.Duration(1) * time.Minute,
			RollingStatsNumBuckets:      int(int64(time.Minute/time.Millisecond) / 100),
			RollingPercentileDuration:   time.Duration(10) * time.Minute,
			RollingPercentileNumBuckets: int(int64(time.Minute/time.Millisecond) / 1000),
			RollingPercentileBucketSize: 50 * 3,
		},

		opt: []cleanhttp.TransportOptions{},
	}

	for _, o := range opt {
		o(params)
	}
	client := cleanhttp.DefaultClient(params.opt...)

	// start the builder with drainer if we have pooling ( its important to drain connections to be able to reclaim them )
	builder := chains.NewChain(params.logger, middlewares.Websocket(), middlewares.Enable(params.pooling, middlewares.Drainer()))

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
			middlewares.NewCircuitMiddleware(params.closerConfig, params.openerConfig, params.statsConfig, params.stream, params.maxConcurrentRequests).Middleware(),
		),
	)

	return builder.ThenClient(client, false)
}
