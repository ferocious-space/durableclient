package middlewares

import (
	"net/http"

	"github.com/ferocious-space/durableclient/chains"
	"github.com/go-logr/logr"
)

func Agent(agent string) chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return chains.RoundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				log := logr.FromContext(request.Context()).WithName("agent")
				log.V(1).Info("Agent()")
				request.Header.Set("User-Agent", agent)
				return next.RoundTrip(request)
			})
	}
}