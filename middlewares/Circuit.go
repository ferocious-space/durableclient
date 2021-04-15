package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/cep21/circuit/metrics/rolling"
	"github.com/cep21/circuit/v3"
	"github.com/cep21/circuit/v3/closers/hystrix"
	"github.com/ferocious-space/durableclient/chains"
	"github.com/go-logr/logr"
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
		crc, err = getManager().CreateCircuit(name)
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
	oneManager.Do(func() {
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
	})
	return m
}

func NewCircuitMiddleware() *CircuitMiddleware {
	return &CircuitMiddleware{}
}

func (c CircuitMiddleware) Middleware(maxErrors int64, rollingDuration time.Duration) chains.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return chains.RoundTripFunc(func(request *http.Request) (*http.Response, error) {
			log := logr.FromContext(request.Context()).WithName("circuit")
			var err error
			crc, err := circuits.Get(request.Host)
			if err != nil {
				return nil, err
			}
			result := new(http.Response)
			err = crc.Execute(request.Context(), func(ctx context.Context) error {
				rsp, err := next.RoundTrip(request.Clone(ctx))
				result = rsp
				return err
			}, func(ctx context.Context, err error) error {
				return nil
			})
			config := crc.ClosedToOpen.(*hystrix.Opener).Config()
			if config.RequestVolumeThreshold != maxErrors {
				crc.ClosedToOpen.(*hystrix.Opener).SetConfigThreadSafe(hystrix.ConfigureOpener{
					RequestVolumeThreshold: maxErrors,
				})
			}
			if config.RollingDuration != rollingDuration {
				crc.ClosedToOpen.(*hystrix.Opener).SetConfigThreadSafe(hystrix.ConfigureOpener{
					RollingDuration: rollingDuration,
				})
			}
			rs := stats.RunStats(request.Host)
			rspBin, _ := httputil.DumpResponse(result, false)
			fmt.Println(string(rspBin))
			log.V(1).Info("Circuit", "Name", crc.Name(), "URI", request.URL.RequestURI(), "L95p", rs.Latencies.Snapshot().Percentile(0.95), "E", rs.ErrFailures.RollingSum(), "T", rs.ErrTimeouts.RollingSum())
			return result, err

			// return next.RoundTrip(request)
		})
	}
}
