package middlewares

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/ferocious-space/durableclient/httpware"
)

func Drainer() httpware.Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		return httpware.RoundTripFunc(func(request *http.Request) (*http.Response, error) {
			rsp, err := next.RoundTrip(request)
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
