package middlewares

import (
	"context"
	"net/http"

	"github.com/ferocious-space/durableclient/chains"
	"github.com/ferocious-space/httpcache"
)

func newCache(cache httpcache.Cache) chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		transport := httpcache.NewTransport(cache)
		transport.Transport = next
		return chains.RoundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				return transport.RoundTrip(request)
			})
	}
}

func Cache(ctx context.Context, cache httpcache.Cache) chains.Chain {
	return chains.NewChain(Drainer(ctx), newCache(cache))
}
