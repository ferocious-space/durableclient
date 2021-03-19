package middlewares

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/improbable-eng/go-httpwares"
)

func AgentDrainer(agent string) httpwares.Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		return httpwares.RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
			request.Header.Set("User-Agent", agent)
			rsp, err := next.RoundTrip(request)
			if err != nil {
				return nil, err
			}
			body, err := ioutil.ReadAll(rsp.Body)
			defer rsp.Body.Close()
			if err != nil {
				return nil, err
			}
			buf := bytes.NewBuffer(body)
			rsp.Body = ioutil.NopCloser(buf)
			return rsp, nil
		})
	}
}
