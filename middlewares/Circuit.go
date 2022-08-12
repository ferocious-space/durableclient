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
	"github.com/cep21/circuit/v3/metriceventstream"
	"github.com/cep21/circuit/v3/metrics/rolling"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/ferocious-space/durableclient/chains"
)

var oneManager sync.Once
var m circuit.Manager
var stats rolling.StatFactory
var mGuard sync.Mutex

func (x *CircuitMiddleware) Get(name string) (crc *circuit.Circuit, err error) {
	crc = getManager(x.stream).GetCircuit(name)
	if crc != nil {
		return crc, nil
	}

	cfg := x.cfg
	stats.RunConfig = *x.scfg
	cfg.Merge(stats.CreateConfig(name))
	crc, err = getManager(x.stream).CreateCircuit(name, cfg)
	if err != nil {
		return nil, err
	}

	return crc, nil
}

type CircuitMiddleware struct {
	cfg    circuit.Config
	scfg   *rolling.RunStatsConfig
	stream bool
}

func getManager(stream bool) *circuit.Manager {
	mGuard.Lock()
	defer mGuard.Unlock()

	oneManager.Do(

		func() {
			factory := hystrix.Factory{}
			m = circuit.Manager{
				DefaultCircuitProperties: []circuit.CommandPropertiesConstructor{factory.Configure, stats.CreateConfig},
			}
			if stream {
				stream := metriceventstream.MetricEventStream{
					Manager:      &m,
					TickDuration: 100 * time.Millisecond,
				}
				http.Handle("/hystrix.stream", &stream)
				go func() {
					stream.Start()
				}()
			}
		},
	)
	return &m
}

func NewCircuitMiddleware(
	closerConfig *hystrix.ConfigureCloser,
	openerConfig *hystrix.ConfigureOpener,
	statsConfig *rolling.RunStatsConfig,
	stream bool,
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
		cfg:    cfg,
		scfg:   statsConfig,
		stream: stream,
	}
}

var (
	redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)
	schemeErrorRe    = regexp.MustCompile(`unsupported protocol scheme`)
	ErrClient4xx     = errors.New("client error")
	ErrServer5xx     = errors.New("server error")
)

func (x *CircuitMiddleware) Middleware() chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return chains.RoundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				ctx := request.Context()
				h, log := logr.FromContextOrDiscard(ctx).WithName("circuit").WithCallStackHelper()
				h()
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
								var ua x509.UnknownAuthorityError
								if ok := errors.Is(err, ua); ok {
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
							// client error
							if rsp.StatusCode >= 400 {
								return ErrClient4xx
							}
							if rsp.StatusCode >= 500 {
								return ErrServer5xx
							}
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

				if err != nil {
					// keep client errors counted only in circuit , dont propagate
					if errors.Is(err, ErrClient4xx) || errors.Is(err, ErrServer5xx) {
						log.V(1).Info(
							"Circuit",
							"Name",
							crc.Name(),
							"URI",
							request.URL.RequestURI(),
							"StatusCode",
							rsp.StatusCode,
							"L95p",
							rs.Latencies.Snapshot().Percentile(0.95).String(),
							"E",
							rs.ErrFailures.RollingSum(),
							"T",
							rs.ErrTimeouts.RollingSum(),
							"O",
							crc.IsOpen(),
						)
						return rsp, nil
					}
					log.V(1).Info(
						"Circuit",
						"Name",
						crc.Name(),
						"URI",
						request.URL.RequestURI(),
						"Error",
						err.Error(),
						"L95p",
						rs.Latencies.Snapshot().Percentile(0.95).String(),
						"E",
						rs.ErrFailures.RollingSum(),
						"T",
						rs.ErrTimeouts.RollingSum(),
						"O",
						crc.IsOpen(),
					)
					return nil, err
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
				return rsp, nil
			},
		)
	}
}
