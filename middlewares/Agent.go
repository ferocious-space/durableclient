package middlewares

import (
	"net/http"

	"github.com/go-logr/logr"

	"github.com/ferocious-space/durableclient/chains"
)

func Agent(agent string) chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return chains.RoundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				logr.FromContextOrDiscard(request.Context()).V(1).Info("middleware.Agent().RoundTripper()", "agent", agent)
				request.Header.Set("User-Agent", agent)
				return next.RoundTrip(request)
			},
		)
	}
}
