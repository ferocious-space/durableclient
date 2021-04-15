package middlewares

import (
	"context"
	"net/http"

	"github.com/ferocious-space/durableclient/chains"
	"github.com/ferocious-space/httpcache"
	"github.com/go-logr/logr"
)

func newCache(cache httpcache.Cache) chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		transport := httpcache.NewTransport(cache)
		transport.Transport = next
		return chains.RoundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				log := logr.FromContext(request.Context()).WithName("cache")
				log.V(1).Info("Cache()")
				return transport.RoundTrip(request)
			})
	}
}

func Cache(ctx context.Context, cache httpcache.Cache) chains.Chain {
	return chains.NewChain(Drainer(ctx), newCache(cache))
}
