package middlewares

import (
	"context"
	"crypto/x509"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/ferocious-space/durableclient/chains"
)

func Retrier(maxRetry int) chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return chains.RoundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				request, err := FromRequest(req)
				if err != nil {
					return nil, err
				}
				log := logr.FromContext(request.Context()).WithName("retrier")

				var rsp *http.Response
				var errRT, checkErr error
				var shouldRetry bool
				var attempt int

				for i := 0; ; i++ {
					attempt++

					if request.Body != nil {
						body, err := request.body()
						if err != nil {
							return nil, err
						}
						if c, ok := body.(io.ReadCloser); ok {
							request.Body = c
						} else {
							request.Body = ioutil.NopCloser(body)
						}
					}

					rsp, errRT = next.RoundTrip(request.Request)
					shouldRetry, checkErr = DefaultRetryPolicy(request.Context(), rsp, errRT)

					// no retry
					if !shouldRetry {
						break
					}

					// chec if we reached max retry
					remain := maxRetry - i
					if remain <= 0 {
						break
					}

					// drain the body of response
					if errRT == nil {
						DrainReader(rsp.Body)
					}

					when := DefaultBackoff(time.Duration(300)*time.Millisecond, time.Duration(3)*time.Second, i, rsp)
					log.V(1).Error(errRT, "retry", "in", when, "uri", request.URL.RequestURI())

					select {
					case <-request.Context().Done():
						return nil, request.Context().Err()
					case <-time.After(when):
					}

					httpReq := *request.Request
					request.Request = &httpReq
				}
				if errRT == nil && checkErr == nil && !shouldRetry {
					return rsp, nil
				}

				err = errRT
				if checkErr != nil {
					err = checkErr
				}

				if rsp != nil {
					DrainReader(rsp.Body)
				}

				if err == nil {
					return nil, errors.Errorf("%s %s giveup after %d attempt(s)", req.Method, req.URL, attempt)
				}
				return nil, errors.WithMessagef(err, "%s %s giveup after %d attempt(s)", req.Method, req.URL, attempt)
			},
		)
	}
}

func DefaultRetryPolicy(ctx context.Context, rsp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if err != nil {
		if redirectsErrorRe.MatchString(err.Error()) {
			return false, err
		}
		if schemeErrorRe.MatchString(err.Error()) {
			return false, err
		}
		if _, ok := err.(x509.UnknownAuthorityError); ok {
			return false, err
		}
		return true, nil
	}
	if rsp.StatusCode == http.StatusTooManyRequests {
		return true, nil
	}
	if rsp.StatusCode == 0 || (rsp.StatusCode >= 500 && rsp.StatusCode != 501) {
		return true, errors.WithMessagef(err, "unexpected HTTP status %s", rsp.Status)
	}
	return false, nil
}

func DefaultBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	if resp != nil {
		if resp.StatusCode == http.StatusTooManyRequests {
			if s, ok := resp.Header["Retry-After"]; ok {
				if sleep, err := strconv.ParseInt(s[0], 10, 64); err == nil {
					return time.Second * time.Duration(sleep)
				}
			}
		}
	}

	rand := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	jitter := rand.Float64() * float64(max-min)
	jitterMin := int64(jitter) + int64(min)
	return time.Duration(jitterMin * int64(attemptNum))
}
