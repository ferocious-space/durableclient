package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/go-logr/logr"

	"github.com/ferocious-space/durableclient/chains"
)

func Drainer() chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return chains.RoundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				h, log := logr.FromContextOrDiscard(request.Context()).WithName("drainer").WithCallStackHelper()
				h()
				log.V(2).Info("middleware.Drainer().RoundTripper()")
				req, err := FromRequest(request)
				if err != nil {
					return nil, err
				}
				if req.Body != nil {
					reqBody, err := req.body()
					if err != nil {
						return nil, err
					}
					if c, ok := reqBody.(io.ReadCloser); ok {
						req.Body = c
					} else {
						req.Body = io.NopCloser(reqBody)
					}
				}
				rsp, err := next.RoundTrip(req.Request)
				if err != nil {
					log.Error(err, "roundtrip()")
					return nil, err
				}
				if rsp != nil && rsp.Body != nil && rsp.ContentLength < 1<<20*20 {
					bodyBuffer := new(bytes.Buffer)
					_, err := io.Copy(bodyBuffer, rsp.Body)
					if err != nil {
						log.Error(err, "reading response body")
						return nil, err
					}
					rsp.Body.Close()
					rsp.Body = io.NopCloser(bodyBuffer)
				}
				return rsp, nil
			},
		)
	}
}
