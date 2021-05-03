package middlewares

import (
	"context"
	"crypto/x509"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/cep21/circuit/v3"
	"github.com/cep21/circuit/v3/closers/hystrix"
	"github.com/cep21/circuit/v3/metrics/rolling"
	"github.com/go-logr/logr"

	"github.com/ferocious-space/durableclient/chains"
)

var oneManager sync.Once
var m circuit.Manager
var stats rolling.StatFactory
var mGuard sync.Mutex

func (x *CircuitMiddleware) Get(name string) (crc *circuit.Circuit, err error) {
	crc = getManager().GetCircuit(name)
	if crc != nil {
		return crc, nil
	}

	cfg := x.cfg
	stats.RunConfig = *x.scfg
	cfg.Merge(stats.CreateConfig(name))
	crc, err = getManager().CreateCircuit(name, cfg)
	if err != nil {
		return nil, err
	}

	return crc, nil
}

type CircuitMiddleware struct {
	cfg  circuit.Config
	scfg *rolling.RunStatsConfig
}

func getManager() *circuit.Manager {
	mGuard.Lock()
	defer mGuard.Unlock()

	oneManager.Do(
		func() {
			factory := hystrix.Factory{}
			m = circuit.Manager{
				DefaultCircuitProperties: []circuit.CommandPropertiesConstructor{factory.Configure, stats.CreateConfig},
			}
		},
	)
	return &m
}

func NewCircuitMiddleware(
	closerConfig *hystrix.ConfigureCloser,
	openerConfig *hystrix.ConfigureOpener,
	statsConfig *rolling.RunStatsConfig,
	maxConcurrentRequests int64,
) *CircuitMiddleware {
	cfg := circuit.Config{
		General: circuit.GeneralConfig{
			ClosedToOpenFactory: hystrix.OpenerFactory(*openerConfig),
			OpenToClosedFactory: hystrix.CloserFactory(*closerConfig),
		},
		Execution: circuit.ExecutionConfig{
			Timeout:               time.Second * time.Duration(5),
			MaxConcurrentRequests: maxConcurrentRequests,
		},
		Fallback: circuit.FallbackConfig{
			MaxConcurrentRequests: maxConcurrentRequests,
		},
		Metrics: circuit.MetricsCollectors{
			Run:      []circuit.RunMetrics{},
			Fallback: []circuit.FallbackMetrics{},
		},
	}

	return &CircuitMiddleware{
		cfg:  cfg,
		scfg: statsConfig,
	}
}

var (
	redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)
	schemeErrorRe    = regexp.MustCompile(`unsupported protocol scheme`)
)

func (x *CircuitMiddleware) Middleware() chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return chains.RoundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				ctx := request.Context()
				log := logr.FromContextOrDiscard(ctx).WithName("circuit")
				log.V(2).Info("middleware.Circuit().RoundTripper()")

				var err error
				crc, err := x.Get(request.Host)
				if err != nil {
					return nil, err
				}

				//configCtO := crc.ClosedToOpen.(*hystrix.Opener).Config()
				//configOtC := crc.OpenToClose.(*hystrix.Closer).Config()

				rs := stats.RunStats(request.Host)
				rsp := new(http.Response)
				err = crc.Run(
					ctx, func(ctx context.Context) error {
						var err error
						hsResp := make(chan *http.Response, 1)
						hsErr := make(chan error, 1)
						go func() {
							defer close(hsResp)
							defer close(hsErr)
							rsp, err = next.RoundTrip(request)
							if err != nil {
								if redirectsErrorRe.MatchString(err.Error()) {
									hsErr <- &circuit.SimpleBadRequest{
										Err: err,
									}
								}
								if schemeErrorRe.MatchString(err.Error()) {
									hsErr <- &circuit.SimpleBadRequest{
										Err: err,
									}
								}
								if _, ok := err.(x509.UnknownAuthorityError); ok {
									hsErr <- &circuit.SimpleBadRequest{
										Err: err,
									}
								}
								hsErr <- err
							}
							hsResp <- rsp
						}()
						select {
						case r := <-hsResp:
							rsp = r
							return nil
						case e := <-hsErr:
							rsp = nil
							return e
						case <-ctx.Done():
							rsp = nil
							return ctx.Err()
						}
					},
				)
				if rsp != nil {
					rsp.Request = request.WithContext(ctx)
				}
				log.V(1).Info(
					"Circuit",
					"Name",
					crc.Name(),
					"URI",
					request.URL.RequestURI(),
					"L95p",
					rs.Latencies.Snapshot().Percentile(0.95).String(),
					"E",
					rs.ErrFailures.RollingSum(),
					"T",
					rs.ErrTimeouts.RollingSum(),
					"O",
					crc.IsOpen(),
				)
				if err != nil {
					return nil, err
				}
				return rsp, nil
			},
		)
	}
}
