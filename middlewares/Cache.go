package middlewares

import (
	"net/http"

	"github.com/ferocious-space/httpcache"

	"github.com/ferocious-space/durableclient/httpware"
)

func Cache(cache httpcache.Cache) httpware.Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		transport := httpcache.NewTransport(cache)
		transport.Transport = next
		return transport
	}
}
