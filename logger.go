package durableclient

import (
	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
)

type privateLogger struct {
	*zap.Logger
}

func newLogger(logger *zap.Logger) retryablehttp.LeveledLogger {
	return &privateLogger{Logger: logger.WithOptions(zap.AddCallerSkip(10))}
}

func (l *privateLogger) Error(msg string, keysAndValues ...interface{}) {
	l.Logger.Error(msg, l.handleFields(keysAndValues)...)
}

func (l *privateLogger) Info(msg string, keysAndValues ...interface{}) {
	l.Logger.Info(msg, l.handleFields(keysAndValues)...)
}

func (l *privateLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.Logger.Debug(msg, l.handleFields(keysAndValues)...)
}

func (l *privateLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.Logger.Warn(msg, l.handleFields(keysAndValues)...)
}

func (l *privateLogger) handleFields(args []interface{}, additional ...zap.Field) []zap.Field {
	// a slightly modified version of zap.SugaredLogger.sweetenFields
	if len(args) == 0 {
		// fast-return if we have no suggared fields.
		return additional
	}

	// unlike Zap, we can be pretty sure users aren't passing structured
	// fields (since logr has no concept of that), so guess that we need a
	// little less space.
	fields := make([]zap.Field, 0, len(args)/2+len(additional))
	for i := 0; i < len(args); {
		// check just in case for strongly-typed Zap fields, which is illegal (since
		// it breaks implementation agnosticism), so we can give a better error message.
		if _, ok := args[i].(zap.Field); ok {
			l.DPanic("strongly-typed Zap Field passed to privateLogger", zap.Any("zap field", args[i]))
			break
		}

		// make sure this isn't a mismatched key
		if i == len(args)-1 {
			l.DPanic("odd number of arguments passed as key-value pairs for logging", zap.Any("ignored key", args[i]))
			break
		}

		// process a key-value pair,
		// ensuring that the key is a string
		key, val := args[i], args[i+1]
		keyStr, isString := key.(string)
		if !isString {
			// if the key isn't a string, DPanic and stop logging
			l.DPanic("non-string key argument passed to logging, ignoring all later arguments", zap.Any("invalid key", key))
			break
		}

		fields = append(fields, zap.Any(keyStr, val))
		i += 2
	}

	return append(fields, additional...)
}
