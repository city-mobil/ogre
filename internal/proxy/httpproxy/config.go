package httpproxy

import (
	"flag"
	"time"
)

type Config struct {
	Port           string
	IdleTimeout    time.Duration
	ReadTimeout    time.Duration
	ErrorThreshold time.Duration
}

func NewConfig() func() *Config {
	var (
		port           = flag.String("httpproxy.port", ":80", "HTTPProxy port")
		idleTimeout    = flag.Duration("httpproxy.server.idle_timeout", 5*time.Second, "HTTPProxy HTTP idle timeout")
		readTimeout    = flag.Duration("httpproxy.server.read_timeout", time.Second, "HTTPProxy HTTP read timeout")
		errorThreshold = flag.Duration("httpproxy.server.error_threshold", 3*time.Minute, "Threshold for error stats")
	)
	return func() *Config {
		return &Config{
			Port:           *port,
			IdleTimeout:    *idleTimeout,
			ReadTimeout:    *readTimeout,
			ErrorThreshold: *errorThreshold,
		}
	}
}
