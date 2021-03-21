package durableclient

import (
	"net/http"

	"github.com/ferocious-space/httpcache"
	"go.uber.org/zap"

	"github.com/ferocious-space/durableclient/httpware"
	"github.com/ferocious-space/durableclient/middlewares"
)

func NewCachedClient(agent string, cache httpcache.Cache, logger *zap.Logger) *http.Client {
	client := httpware.CloneClient()
	client.Logger = newLogger(logger.WithOptions(zap.AddCallerSkip(13)))
	return httpware.TripperwareStack(middlewares.Drainer(), middlewares.Cache(cache), middlewares.Agent(agent)).DecorateClient(client.StandardClient(), false)
}

func NewClient(agent string, logger *zap.Logger) *http.Client {
	client := httpware.CloneClient()
	client.Logger = newLogger(logger.WithOptions(zap.AddCallerSkip(12)))
	return httpware.TripperwareStack(middlewares.Drainer(), middlewares.Agent(agent)).DecorateClient(client.StandardClient(), false)
}
