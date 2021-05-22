package middlewares

import (
	"net/http"

	"github.com/go-logr/logr"

	"github.com/ferocious-space/durableclient/chains"
)

func Logger(log logr.Logger) chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return chains.RoundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				return next.RoundTrip(request.Clone(logr.NewContext(request.Context(), logr.WithCallDepth(log, 7))))
			},
		)
	}
}
