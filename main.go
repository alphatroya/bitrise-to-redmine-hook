package main

import (
	"context"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/alphatroya/ci-redmine-bindings/settings"
	"github.com/getsentry/sentry-go"
	"github.com/go-redis/redis"
	"github.com/rs/zerolog"
)

func init() {
	zerolog.ErrorFieldName = "err"

	buildInfo, _ := debug.ReadBuildInfo()
	logger := zerolog.
		New(os.Stdout).
		Level(zerolog.TraceLevel).
		With().
		Timestamp().
		Caller().
		Int("pid", os.Getpid()).
		Str("go_version", buildInfo.GoVersion).
		Logger()

	zerolog.DefaultContextLogger = &logger
}

func main() {
	logger := zerolog.Ctx(context.Background())

	settings, err := settings.Current()
	if err != nil {
		logger.Fatal().
			Err(err).
			Msg("fail to collect required env configuration")
	}

	if err = sentry.Init(sentry.ClientOptions{
		Dsn:              settings.SentryDSN,
		TracesSampleRate: 1.0,
	}); err != nil {
		logger.Fatal().
			Err(err).
			Msg("sentry init failed")
	}
	defer sentry.Flush(2 * time.Second)

	stamper, err := createStamper(settings)
	if err != nil {
		logger.Fatal().
			Err(err).
			Msg("failed to create handler instance")
	}
	http.Handle("/bitrise", stamper)
	http.Handle("/bitrise/v2", stamper)
	//nolint
	if err := http.ListenAndServe(":"+settings.Port, nil); err != nil {
		logger.Fatal().
			Err(err).
			Msg("can't open up the server")
	}
}

func createStamper(settings *settings.Config) (*Stamper, error) {
	options, err := redis.ParseURL(settings.RedisURL)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(options)
	_, err = rdb.Ping().Result()
	if err != nil {
		return nil, err
	}
	return NewStamper(settings, rdb), nil
}
