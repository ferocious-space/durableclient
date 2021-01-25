package durableclient

import (
	"bytes"
	"io/ioutil"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gojektech/heimdall/v6"
	"github.com/gojektech/heimdall/v6/httpclient"
	"github.com/gojektech/heimdall/v6/hystrix"
	"go.uber.org/zap"
)

var registryMU sync.RWMutex
var clientRegistry map[string]*http.Client

type rClient struct {
	client     *hystrix.Client
	agent      string
	transports []*http.Transport
}

type innerClient struct {
	heimdall.Doer
	logger *zap.Logger
}

func (r *innerClient) Do(req *http.Request) (rsp *http.Response, err error) {
	rsp, err = r.Doer.Do(req)
	if err != nil {
		r.logger.Error(req.URL.String(), zap.Error(err))
	}
	if err == nil && rsp != nil {
		if rsp.StatusCode >= 500 {
			r.logger.Error(rsp.Request.URL.String(), zap.Int("StatusCode", rsp.StatusCode))
			r.logger.Debug(rsp.Request.URL.String(), zap.Any("Request Headers", rsp.Request.Header))
			r.logger.Debug(rsp.Request.URL.String(), zap.Any("Headers", rsp.Header))
		}
		if rsp.StatusCode >= 400 && rsp.StatusCode < 500 {
			r.logger.Warn(rsp.Request.URL.String(), zap.Int("StatusCode", rsp.StatusCode))
			r.logger.Debug(rsp.Request.URL.String(), zap.Any("Request Headers", rsp.Request.Header))
			r.logger.Debug(rsp.Request.URL.String(), zap.Any("Headers", rsp.Header))
		}
		if rsp.StatusCode > 200 && rsp.StatusCode < 400 {
			r.logger.Info(rsp.Request.URL.String(), zap.Int("StatusCode", rsp.StatusCode))
		}
		if rsp.StatusCode == 200 {
			r.logger.Debug(rsp.Request.URL.String(), zap.Int("StatusCode", rsp.StatusCode))
		}
	}
	return rsp, err
}

func NewClient(name string, agent string, logger *zap.Logger) *http.Client {
	registryMU.Lock()
	defer registryMU.Unlock()
	if clientRegistry == nil {
		clientRegistry = make(map[string]*http.Client)
	}
	if c, OK := clientRegistry[name]; OK {
		return c
	} else {
		clientRegistry[name] = &http.Client{
			Transport: &rClient{
				agent: agent,
				client: hystrix.NewClient(
					hystrix.WithHTTPTimeout(300*time.Millisecond),
					hystrix.WithHystrixTimeout(1000*time.Millisecond),
					hystrix.WithCommandName(name),
					// stop requets if 5% of the fail which is around 50 requests
					hystrix.WithErrorPercentThreshold(5),
					// max concurrent requets of type
					hystrix.WithMaxConcurrentRequests(1000),
					hystrix.WithRetryCount(3),
					hystrix.WithRetrier(
						heimdall.NewRetrier(
							heimdall.NewExponentialBackoff(200*time.Millisecond, 5*time.Second, math.SqrtPi, 10*time.Millisecond),
						),
					),
					// time to sleep when circuit open - 1 min
					hystrix.WithSleepWindow(60000),
					hystrix.WithHTTPClient(&innerClient{
						Doer:   httpclient.NewClient(),
						logger: logger.Named("hystrix"),
					}),
				),
			},
		}
	}
	return clientRegistry[name]
}

func (r *rClient) GetClient() *http.Client {
	return &http.Client{
		Transport: r,
	}
}

func (r *rClient) RoundTrip(req *http.Request) (*http.Response, error) {
	if r.agent != "" {
		req.Header.Set("User-Agent", r.agent)
	}
	rsp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewReader(body)
	rsp.Body = ioutil.NopCloser(buf)
	return rsp, nil
}
