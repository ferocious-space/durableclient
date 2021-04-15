package middlewares

import (
	"net/http"

	"github.com/ferocious-space/durableclient/chains"
)

func Retrier() chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return chains.RoundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				return next.RoundTrip(request)
			})
	}
}
