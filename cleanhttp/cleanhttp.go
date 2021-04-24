package cleanhttp

// taken from hashicorp/go-cleanhttp
import (
	"crypto/tls"
	"net"
	"net/http"
	"runtime"
	"time"
)

func DefaultTransport(opt ...TransportOptions) *http.Transport {
	params := &clientParams{
		http2disabled: false,
		pooledclient:  false,
	}

	for _, o := range opt {
		o(params)
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if !params.pooledclient {
		transport.DisableKeepAlives = true
		transport.MaxIdleConnsPerHost = -1
	} else {
		transport.MaxIdleConnsPerHost = runtime.GOMAXPROCS(0) + 1
	}

	if params.http2disabled {
		transport.TLSNextProto = map[string]func(authority string, c *tls.Conn) http.RoundTripper{}
	}

	return transport
}

func DefaultClient(opt ...TransportOptions) *http.Client {
	return &http.Client{
		Transport: DefaultTransport(opt...),
	}
}
