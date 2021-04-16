package middlewares

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ferocious-space/durableclient/chains"
	"github.com/go-logr/logr"
)

func Drainer(ctx context.Context) chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return chains.RoundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				log := logr.FromContext(ctx).WithName("drainer")
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
				rsp, err := next.RoundTrip(request.WithContext(ctx))
				if err != nil {
					log.Error(err, "roundtrip()")
					return nil, err
				}
				if rsp != nil && rsp.Body != nil {
					body, err := ioutil.ReadAll(rsp.Body)
					if err != nil {
						log.Error(err, "reading response body")
						return nil, err
					}
					rsp.Body.Close()
					rsp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
				}
				return rsp, nil
			})
	}
}
