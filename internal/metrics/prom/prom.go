package prom

import "github.com/city-mobil/gobuns/promlib"

const (
	SentryRequestCount       = "sentry_request_total"
	SentryFailedRequestCount = "sentry_failed_request_total"
	OgreQueuePutError        = "failed_put_to_queue_total"
	OgrePutToQueue           = "put_to_queue_total"
	OgreReqCnt               = "req_count"
)

var (
	OgreQueueRatio = promlib.NewGauge(promlib.GaugeOptions{
		MetaOpts: promlib.MetaOpts{
			Name: "queue_size_ratio",
		},
	})
)
