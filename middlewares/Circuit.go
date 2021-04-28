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
	"github.com/pkg/errors"

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

func (c CircuitMiddleware) Middleware(maxErrors int64, ErrorThresholdPercentage int64, rollingDuration time.Duration) chains.Middleware {
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

				config := crc.ClosedToOpen.(*hystrix.Opener).Config()
				if config.RequestVolumeThreshold != maxErrors {
					config.Merge(
						hystrix.ConfigureOpener{
							RequestVolumeThreshold: maxErrors,
						},
					)
					crc.ClosedToOpen.(*hystrix.Opener).SetConfigThreadSafe(
						config,
					)
				}
				if config.RollingDuration != rollingDuration {
					config.Merge(
						hystrix.ConfigureOpener{
							RollingDuration: rollingDuration,
						},
					)
					crc.ClosedToOpen.(*hystrix.Opener).SetConfigThreadSafe(
						config,
					)
				}

				if config.ErrorThresholdPercentage != ErrorThresholdPercentage {
					config.Merge(
						hystrix.ConfigureOpener{
							ErrorThresholdPercentage: ErrorThresholdPercentage,
						},
					)
					crc.ClosedToOpen.(*hystrix.Opener).SetConfigThreadSafe(
						config,
					)
				}

				result := make(chan *http.Response, 1)
				defer close(result)
				err = crc.Execute(
					request.Context(), func(ctx context.Context) error {
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
				rs := stats.RunStats(request.Host)
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
				)
				if err != nil {
					return nil, err
				}
				select {
				case x := <-result:
					return x, nil
				case <-time.After(1 * time.Second):
					return nil, errors.New("timeout waiting for reply")
				}
			},
		)
	}
}
