package middlewares

import (
	"net/http"

	"github.com/ferocious-space/durableclient/chains"
)

func Enable(enable bool, middleware chains.Middleware) chains.Middleware {
	if enable {
		return middleware
	}
	return func(next http.RoundTripper) http.RoundTripper {
		return next
	}
}
