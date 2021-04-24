package durableclient

import (
	"context"

	"github.com/ferocious-space/httpcache"
	"github.com/go-logr/logr"
)

type durableClientOptions struct {
	ctx     context.Context
	logger  logr.Logger
	agent   string
	cache   httpcache.Cache
	pooled  bool
	retrier bool
	http2   bool
}
