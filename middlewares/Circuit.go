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
var m *circuit.Manager
var stats rolling.StatFactory

var circuits circuitMap

type circuitMap struct {
	m sync.Map
}

func (m *circuitMap) Get(name string) (crc *circuit.Circuit, err error) {
	if data, ok := m.m.Load(name); !ok {
		crc, err = getManager().CreateCircuit(
			name, circuit.Config{
				Execution: circuit.ExecutionConfig{
					Timeout:               -1,
					MaxConcurrentRequests: 50,
				},
			},
		)
		if err != nil {
			return nil, err
		}
		m.m.Store(name, crc)
	} else {
		if crc, ok = data.(*circuit.Circuit); !ok {
			crc, err = getManager().CreateCircuit(name)
			if err != nil {
				return nil, err
			}
			m.m.Store(name, crc)
		}
	}
	return crc, nil
}

type CircuitMiddleware struct {
}

func getManager() *circuit.Manager {
	oneManager.Do(
		func() {
			factory := hystrix.Factory{
				ConfigureCloser: hystrix.ConfigureCloser{
					SleepWindow:                  time.Minute * 5,
					HalfOpenAttempts:             10,
					RequiredConcurrentSuccessful: 3,
				},
				ConfigureOpener: hystrix.ConfigureOpener{
					ErrorThresholdPercentage: 80,
					RequestVolumeThreshold:   80,
					RollingDuration:          time.Second * 60,
				},
			}
			m = &circuit.Manager{
				DefaultCircuitProperties: []circuit.CommandPropertiesConstructor{factory.Configure, stats.CreateConfig},
			}
		},
	)
	return m
}

func NewCircuitMiddleware() *CircuitMiddleware {
	return &CircuitMiddleware{}
}

var (
	redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)
	schemeErrorRe    = regexp.MustCompile(`unsupported protocol scheme`)
)

func (c CircuitMiddleware) Middleware(maxErrors int64, rollingDuration time.Duration) chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return chains.RoundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				logr.FromContextOrDiscard(request.Context()).V(2).Info("middleware.Circuit().RoundTripper()", "maxErrors", maxErrors, "rollingWindow", rollingDuration)
				log := logr.FromContextOrDiscard(request.Context()).WithName("circuit")
				var err error
				crc, err := circuits.Get(request.Host)
				if err != nil {
					return nil, err
				}

				result := make(chan *http.Response, 1)
				err = crc.Execute(
					request.Context(), func(ctx context.Context) error {
						defer close(result)
						rsp, err := next.RoundTrip(request.Clone(ctx))
						result <- rsp
						if err != nil {
							if redirectsErrorRe.MatchString(err.Error()) {
								return &circuit.SimpleBadRequest{
									Err: err,
								}
							}
							if schemeErrorRe.MatchString(err.Error()) {
								return &circuit.SimpleBadRequest{
									Err: err,
								}
							}
							if _, ok := err.(x509.UnknownAuthorityError); ok {
								return &circuit.SimpleBadRequest{
									Err: err,
								}
							}
							return err
						}
						return nil
					}, nil,
				)
				config := crc.ClosedToOpen.(*hystrix.Opener).Config()
				if config.RequestVolumeThreshold != maxErrors {
					crc.ClosedToOpen.(*hystrix.Opener).SetConfigThreadSafe(
						hystrix.ConfigureOpener{
							RequestVolumeThreshold: maxErrors,
						},
					)
				}
				if config.RollingDuration != rollingDuration {
					crc.ClosedToOpen.(*hystrix.Opener).SetConfigThreadSafe(
						hystrix.ConfigureOpener{
							RollingDuration: rollingDuration,
						},
					)
				}
				rs := stats.RunStats(request.Host)
				res := <-result
				log.V(1).Info(
					"Circuit",
					"Name",
					crc.Name(),
					"URI",
					request.URL.RequestURI(),
					"L95p",
					rs.Latencies.Snapshot().Percentile(0.95),
					"E",
					rs.ErrFailures.RollingSum(),
					"T",
					rs.ErrTimeouts.RollingSum(),
				)
				return res, err

				// return next.RoundTrip(request)
			},
		)
	}
}
