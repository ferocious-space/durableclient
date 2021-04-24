package durableclient

import (
	"context"
	"time"

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
	numRetries int

	circuit       bool
	maxErrors     int64
	rollingWindow time.Duration

	opt []cleanhttp.TransportOptions
}

type ClientOptions func(c *durableOption)

// OptionAgent changes default agent
func OptionAgent(agent string) ClientOptions {
	return func(c *durableOption) {
		c.agent = agent
	}
}

// OptionLogger adds default logger to the request context
func OptionLogger(logger logr.Logger) ClientOptions {
	return func(c *durableOption) {
		c.ctx = logr.NewContext(c.ctx, logger)
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
func OptionCircuit(maxErrors int64, rollingWindow time.Duration) ClientOptions {
	return func(c *durableOption) {
		if maxErrors <= 0 || rollingWindow <= time.Duration(0) {
			c.circuit = false
			return
		}
		c.circuit = true
		c.maxErrors = maxErrors
		c.rollingWindow = rollingWindow
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
