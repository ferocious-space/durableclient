package middlewares

import (
	"net/http"

	"github.com/go-logr/logr"

	"github.com/ferocious-space/durableclient/chains"
	"github.com/ferocious-space/durableclient/cleanhttp"
)

func Websocket() chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return chains.RoundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				if request.Header.Get("Upgrade") == "websocket" && request.Header.Get("Connection") == "Upgrade" {
					h, l := logr.FromContextOrDiscard(request.Context()).WithCallStackHelper()
					h()
					l.V(2).Info("middleware.Websocket().RoundTripper()")
					return cleanhttp.DefaultClient().Transport.RoundTrip(request)
				}
				return next.RoundTrip(request)
			},
		)
	}
}
