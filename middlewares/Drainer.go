package middlewares

import (
	"bytes"
	"context"
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
				log.V(1).Info("Drainer()")
				if request.Body != nil {
					reqBody, err := ioutil.ReadAll(request.Body)
					if err != nil {
						request.Body.Close()
						return nil, err
					}
					request.Body.Close()
					request.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

				}
				rsp, err := next.RoundTrip(request.Clone(ctx))
				if err != nil {
					return nil, err
				}
				body, err := ioutil.ReadAll(rsp.Body)
				if err != nil {
					return nil, err
				}
				rsp.Body.Close()
				rsp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
				return rsp, nil
			})
	}
}
