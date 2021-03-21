package middlewares

import (
	"net/http"

	"github.com/ferocious-space/durableclient/httpware"
)

func Agent(agent string) httpware.Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		return httpware.RoundTripFunc(func(request *http.Request) (*http.Response, error) {
			request.Header.Set("User-Agent", agent)
			return next.RoundTrip(request)
		})
	}
}
