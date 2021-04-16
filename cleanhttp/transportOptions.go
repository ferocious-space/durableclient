package cleanhttp

import (
	"crypto/tls"
	"net/http"
)

type TransportOptions func(t *http.Transport)

func WithHTTP2Disabled() TransportOptions {
	return func(t *http.Transport) {
		t.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
	}
}
