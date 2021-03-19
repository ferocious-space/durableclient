package durableclient

import (
	"net/http"
	"net/http/cookiejar"

	"github.com/ferocious-space/httpcache"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/improbable-eng/go-httpwares"
	"go.uber.org/zap"
	"golang.org/x/net/publicsuffix"

	"github.com/ferocious-space/durableclient/middlewares"
)

func NewCachedClient(agent string, cache httpcache.Cache, logger *zap.Logger) *http.Client {
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil
	}

	httpClient := retryablehttp.NewClient()
	httpClient.Logger = newLogger(logger.WithOptions(zap.AddCallerSkip(1)))

	httpClient.Backoff = retryablehttp.LinearJitterBackoff
	httpClient.RetryMax = 3
	httpClient.RetryWaitMin = 300
	httpClient.RetryWaitMax = 700
	httpClient.HTTPClient.Jar = jar

	cacheTransport := httpcache.NewTransport(cache)
	cacheTransport.Transport = httpClient.StandardClient().Transport
	return httpwares.WrapClient(cacheTransport.Client(), middlewares.AgentDrainer(agent))
}

func NewClient(agent string, logger *zap.Logger) *http.Client {
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil
	}

	httpClient := retryablehttp.NewClient()
	httpClient.Logger = newLogger(logger)

	httpClient.Backoff = retryablehttp.LinearJitterBackoff
	httpClient.RetryMax = 3
	httpClient.RetryWaitMin = 300
	httpClient.RetryWaitMax = 700
	httpClient.HTTPClient.Jar = jar
	return httpwares.WrapClient(httpClient.StandardClient(), middlewares.AgentDrainer(agent))
}
