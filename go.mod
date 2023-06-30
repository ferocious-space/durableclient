module github.com/ferocious-space/durableclient

go 1.20

retract [v0.0.0, v0.7.3]

require (
	github.com/cep21/circuit/v3 v3.2.2
	github.com/ferocious-space/httpcache v0.0.0-20230630110858-8f77f1862b7a
	github.com/go-logr/logr v1.2.4
	github.com/go-logr/zapr v1.2.4
	github.com/pkg/errors v0.9.1
	go.uber.org/zap v1.24.0
	golang.org/x/net v0.11.0
)

require (
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
)
