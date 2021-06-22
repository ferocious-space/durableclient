package durableclient

import (
	"context"
	"time"

	"github.com/cep21/circuit/v3/closers/hystrix"
	"github.com/cep21/circuit/v3/metrics/rolling"
	"github.com/ferocious-space/httpcache"
	"github.com/go-logr/logr"

	"github.com/ferocious-space/durableclient/cleanhttp"
)

type durableOption struct {
	cache      httpcache.Cache
	ctx        context.Context
	logger     logr.Logger
	agent      string
	pooling    bool
	retrier    bool
	stream     bool
	numRetries int

	circuit               bool
	maxConcurrentRequests int64
	closerConfig          *hystrix.ConfigureCloser
	openerConfig          *hystrix.ConfigureOpener
	statsConfig           *rolling.RunStatsConfig

	opt []cleanhttp.TransportOptions
}

type ClientOptions func(c *durableOption)

// OptionAgent changes default agent
func OptionAgent(agent string) ClientOptions {
	return func(c *durableOption) {
		c.agent = agent
	}
}

// OptionStream adds hystrix even stream to /hystrix.stream http.defaultservermux
func OptionStream() ClientOptions {
	return func(c *durableOption) {
		c.stream = true
	}
}

// OptionLogger adds default logger to the request context
func OptionLogger(logger logr.Logger) ClientOptions {
	return func(c *durableOption) {
		c.logger = logr.WithCallDepth(logger, 1)
	}
}

// OptionCache adds cache to the client
func OptionCache(cache httpcache.Cache) ClientOptions {
	return func(c *durableOption) {
		c.cache = cache
	}
}

// OptionConnectionPooling enables http.Client connection pooling and reuse
// this options enables middlewares.Drainer and modifies http.Transport
func OptionConnectionPooling(enable bool) ClientOptions {
	return func(c *durableOption) {
		c.pooling = enable
		if enable {
			c.opt = append(c.opt, cleanhttp.WithPooledClient())
		}
	}
}

//OptionRetrier if numRetries <= 0 , disables the retry mechanisms
func OptionRetrier(numRetries int) ClientOptions {
	return func(c *durableOption) {
		if numRetries <= 0 {
			c.retrier = false
			return
		}
		c.retrier = true
		c.numRetries = numRetries
	}
}

//OptionCircuit if maxErrors <= or rollingWindow <= 0 , disables circuit
func OptionCircuit(maxErrors int64, ErrorThresholdPercentage int64, rollingWindow time.Duration, sleepWindow time.Duration, maxConcurrentRequests int64) ClientOptions {
	return func(c *durableOption) {
		if maxErrors <= 0 || rollingWindow <= time.Duration(0) {
			c.circuit = false
			return
		}
		c.circuit = true
		c.maxConcurrentRequests = maxConcurrentRequests

		c.openerConfig.RequestVolumeThreshold = maxErrors
		c.openerConfig.ErrorThresholdPercentage = ErrorThresholdPercentage

		c.openerConfig.RollingDuration = rollingWindow
		c.openerConfig.NumBuckets = int(int64(rollingWindow/time.Millisecond) / 100)

		c.closerConfig.SleepWindow = sleepWindow
		c.closerConfig.HalfOpenAttempts = int64(sleepWindow/time.Second) / 5
		c.closerConfig.RequiredConcurrentSuccessful = 3

		c.statsConfig.RollingStatsDuration = c.openerConfig.RollingDuration
		c.statsConfig.RollingStatsNumBuckets = int(int64(c.statsConfig.RollingStatsDuration/time.Millisecond) / 100)
		c.statsConfig.RollingPercentileDuration = c.openerConfig.RollingDuration * 10
		c.statsConfig.RollingPercentileNumBuckets = int(int64(c.statsConfig.RollingPercentileDuration/time.Millisecond) / 1000)
		c.statsConfig.RollingPercentileBucketSize = int(maxConcurrentRequests * 3)
	}
}

//HTTP2 is enabled by default in http.Transport
func OptionHTTP2(enabled bool) ClientOptions {
	return func(c *durableOption) {
		if !enabled {
			c.opt = append(c.opt, cleanhttp.WithHTTP2Disabled())
		}
	}
}
