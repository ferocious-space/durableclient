package cleanhttp

// taken from hashicorp/go-cleanhttp
import (
	"crypto/tls"
	"net"
	"net/http"
	"runtime"
	"time"
)

// DefaultTransport returns a new http.Transport with similar default values to
// http.DefaultTransport, but with idle connections and keepalives disabled.
func DefaultTransport(disableHTTP2 ...bool) *http.Transport {
	transport := DefaultPooledTransport(disableHTTP2...)
	transport.DisableKeepAlives = true
	transport.MaxIdleConnsPerHost = -1
	return transport
}

// DefaultPooledTransport returns a new http.Transport with similar default
// values to http.DefaultTransport. Do not use this for transient transports as
// it can leak file descriptors over time. Only use this for transports that
// will be re-used for the same host(s).
func DefaultPooledTransport(disableHTTP2 ...bool) *http.Transport {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSNextProto:          map[string]func(string, *tls.Conn) http.RoundTripper{},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}
	for _, b := range disableHTTP2 {
		if b {
			transport.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
			break
		}
	}
	return transport
}

// DefaultClient returns a new http.Client with similar default values to
// http.Client, but with a non-shared Transport, idle connections disabled, and
// keepalives disabled.
func DefaultClient(disableHTTP2 ...bool) *http.Client {
	return &http.Client{
		Transport: DefaultTransport(disableHTTP2...),
	}
}

// DefaultPooledClient returns a new http.Client with similar default values to
// http.Client, but with a shared Transport. Do not use this function for
// transient clients as it can leak file descriptors over time. Only use this
// for clients that will be re-used for the same host(s).
func DefaultPooledClient(disableHTTP2 ...bool) *http.Client {
	return &http.Client{
		Transport: DefaultPooledTransport(disableHTTP2...),
	}
}
