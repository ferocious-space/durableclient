package middlewares

import (
	"net/http"

	"github.com/ferocious-space/durableclient/chains"
)

func Agent(agent string) chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return chains.RoundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				request.Header.Set("User-Agent", agent)
				return next.RoundTrip(request)
			})
	}
}
