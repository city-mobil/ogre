package pusher

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"ogre/internal/metrics/prom"

	"github.com/city-mobil/gobuns/promlib"
	"github.com/shmel1k/gop"
)

var (
	ErrItemDropped   = errors.New("pusher: event dropped due to timeout")
	ErrPusherStopped = errors.New("pusher: is stopped")
)

type Pusher interface {
	Push(context.Context, Item) error
	Stop()
}

type pusherParams struct {
	hosts           []string
	projectMap      map[string]string
	projectIDregexp *regexp.Regexp
}

type sentryPusher struct {
	config      *Config
	params      *pusherParams
	client      *http.Client
	onceStopper sync.Once
	stop        chan struct{}
	pool        gop.Pool
}

func (p *sentryPusher) Push(ctx context.Context, it Item) error {
	queueRatio := float64(p.pool.QueueSize()) / float64(p.config.QueueSize)
	prom.OgreQueueRatio.Set(queueRatio)
	return p.pool.AddContext(ctx, gop.TaskFn(func() {
		err := p.pushToSentry(it)
		if err != nil {
			log.Printf("failed to perform task %v: %s", it.GetEventID(), err)
		}
	}))
}

func (p *sentryPusher) pushToSentry(it Item) error {
	u := it.GetURL()
	h := it.GetHeader()
	sentryHeader := h.Get("X-Sentry-Auth")
	originalProjectID := p.params.projectIDregexp.FindString(sentryHeader)

	for i, host := range p.params.hosts {
		hostURL, err := url.Parse(host)
		if err != nil {
			log.Printf("error parsing host %s\n", host)
			continue
		}
		u.Host = hostURL.Host
		u.Scheme = hostURL.Scheme
		req, err := http.NewRequest(it.GetMethod(), u.String(), bytes.NewReader(it.GetBody()))
		if err != nil {
			return err
		}

		req.Header = it.GetHeader()
		if i != 0 {
			newProjectID := p.params.projectMap[originalProjectID]
			if newProjectID == "" {
				log.Printf("ERROR: no projectID coupled for %s\n", originalProjectID)
			}
			req.Header.Set("X-Sentry-Auth", p.params.projectIDregexp.ReplaceAllString(sentryHeader, newProjectID))
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			promlib.IncCnt(prom.SentryFailedRequestCount)
			return err
		}
		if resp != nil {
			if err := resp.Body.Close(); err != nil {
				return err
			}
		}
	}

	promlib.IncCnt(prom.SentryRequestCount)
	return nil
}

func (p *sentryPusher) Stop() {
	p.onceStopper.Do(func() {
		close(p.stop)
	})
}

func NewPusher(cfg *Config) Pusher {
	client := &http.Client{
		Timeout: cfg.ClientRequestTimeout,
	}
	pool := gop.NewPool(gop.Config{
		UnstoppableWorkers:  cfg.PoolUnstoppableWorkers,
		TaskScheduleTimeout: cfg.DropThreshold,
		MaxQueueSize:        cfg.QueueSize,
	})

	pusherParams, err := getPusherParams(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	return &sentryPusher{
		config: cfg,
		params: pusherParams,
		client: client,
		pool:   pool,
		stop:   make(chan struct{}),
	}
}

func getPusherParams(config *Config) (*pusherParams, error) {
	hosts := strings.Split(config.SentryHost, ",")
	regex := regexp.MustCompile(`\w{32}:\w{32},?`)
	lines := regex.FindAllString(config.ProjectMap, -1)

	m := make(map[string]string)
	for _, line := range lines {
		projectIDs := strings.Split(strings.Replace(line, ",", "", -1), ":")
		if len(hosts) != len(projectIDs) {
			return nil, errors.New("count of hosts does not correlate with project map")
		}
		m[projectIDs[0]] = projectIDs[1]
	}

	re := regexp.MustCompile(`(\w{32})`)
	params := &pusherParams{
		hosts:           hosts,
		projectMap:      m,
		projectIDregexp: re,
	}
	return params, nil
}
