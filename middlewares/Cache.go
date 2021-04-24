package middlewares

import (
	"fmt"
	"net/http"

	"github.com/ferocious-space/httpcache"
	"github.com/go-logr/logr"

	"github.com/ferocious-space/durableclient/chains"
)

func Cache(cache httpcache.Cache) chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		transport := httpcache.NewTransport(cache)
		transport.Transport = next
		return chains.RoundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				logr.FromContextOrDiscard(request.Context()).V(1).Info("middleware.Cache().RoundTripper()", "type", fmt.Sprintf("%T", cache))
				return transport.RoundTrip(request)
			},
		)
	}
}
