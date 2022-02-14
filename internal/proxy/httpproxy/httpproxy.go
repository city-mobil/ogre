package httpproxy

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"ogre/internal/metrics/prom"
	"ogre/internal/proxy"
	"ogre/internal/pusher"

	b "github.com/city-mobil/gobuns/barber"
	"github.com/city-mobil/gobuns/promlib"
	"github.com/google/uuid"
)

type httpProxy struct {
	config *Config
	server *http.Server
	psher  pusher.Pusher
	barber b.Barber

	onceStopper sync.Once
}

func New(cfg *Config, psher pusher.Pusher, barber b.Barber) proxy.Proxy {
	server := &http.Server{
		Addr:        cfg.Port,
		IdleTimeout: cfg.IdleTimeout,
		ReadTimeout: cfg.ReadTimeout,
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	return &httpProxy{
		config: cfg,
		server: server,
		psher:  psher,
		barber: barber,
	}
}

func (h *httpProxy) handler(w http.ResponseWriter, r *http.Request) {
	promlib.IncCnt(prom.OgreReqCnt)
	h.barber.AddError(0, time.Now())

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	u, _ := uuid.NewRandom()
	eventID := strings.Replace(u.String(), "-", "", 4)

	_, err := w.Write(getResponse(eventID))
	if err != nil {
		log.Println("Error during send response", err)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("event_id: %s, got request %s", eventID, r.URL)

	p := pusher.NewItem(body, r.Method, r.Header, r.URL, eventID)

	go func() {
		err := h.psher.Push(context.Background(), p)
		if err != nil {
			promlib.IncCnt(prom.OgreQueuePutError)
			log.Println("proxy queue wait timout")
		} else {
			promlib.IncCnt(prom.OgrePutToQueue)
		}
	}()
}

func (h *httpProxy) Start() <-chan error {
	errs := make(chan error, 1)
	promMW := promlib.NewMiddleware(promlib.DefHTTPRequestDurBuckets)

	http.HandleFunc("/", promMW.HandlerFunc(h.handler))
	go func() {
		errs <- h.server.ListenAndServe()
	}()

	return errs
}

func (h *httpProxy) Stop(ctx context.Context) error {
	var err error
	h.onceStopper.Do(func() {
		err = h.server.Shutdown(ctx)
	})
	return err
}

func getResponse(eventID string) []byte {
	resp := make([]byte, 0, 46)
	resp = append(resp, `{"id": "`...)
	resp = append(resp, eventID...)
	resp = append(resp, `"}`...)
	return resp
}
