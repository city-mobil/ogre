package pusher

import (
	"flag"
	"time"
)

type Config struct {
	SentryHost             string
	ProjectMap             string
	DropThreshold          time.Duration
	ClientRequestTimeout   time.Duration
	QueueSize              int
	PoolUnstoppableWorkers int
}

func NewConfig() func() *Config {
	var (
		sentryHost     = flag.String("pusher.sentry.host", "", "Pusher address of sentry hosts (necessarily with scheme), use comma to separate hostnames")
		projectMap     = flag.String("pusher.sentry.projects", "", "Map with projects for sentry. Checkout readme for more info.")
		dropThreshold  = flag.Duration("pusher.drop_threshold", 20*time.Millisecond, "Event drop threshold duration")
		queueSize      = flag.Int("pusher.queue_size", 1000, "Pusher queue size")
		requestTimeout = flag.Duration("pusher.client.request_timeout", 100*time.Millisecond, "Pusher HTTP client timeout")
		poolSize       = flag.Int("pusher.pool.unstoppable_workers", 4, "Pusher pool unstoppable workers count.")
	)

	return func() *Config {
		return &Config{
			SentryHost:             *sentryHost,
			ProjectMap:             *projectMap,
			DropThreshold:          *dropThreshold,
			QueueSize:              *queueSize,
			ClientRequestTimeout:   *requestTimeout,
			PoolUnstoppableWorkers: *poolSize,
		}
	}
}
