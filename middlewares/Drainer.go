package middlewares

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-logr/logr"

	"github.com/ferocious-space/durableclient/chains"
)

func Drainer() chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return chains.RoundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				logr.FromContextOrDiscard(request.Context()).V(2).Info("middleware.Drainer().RoundTripper()")
				log := logr.FromContext(request.Context()).WithName("drainer")
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
						req.Body = ioutil.NopCloser(reqBody)
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
					rsp.Body = ioutil.NopCloser(bodyBuffer)
				}
				return rsp, nil
			},
		)
	}
}
