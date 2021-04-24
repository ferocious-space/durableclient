package cleanhttp

type clientParams struct {
	http2disabled bool
	pooledclient  bool
}

type TransportOptions func(p *clientParams)

func WithHTTP2Disabled() TransportOptions {
	return func(p *clientParams) {
		p.http2disabled = true
	}
}

//WithPooledClient elables connection pooling on the *http.Client
func WithPooledClient() TransportOptions {
	return func(p *clientParams) {
		p.pooledclient = true
	}
}
