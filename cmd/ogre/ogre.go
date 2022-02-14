package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // Import
	"os"
	"os/signal"
	"syscall"
	"time"

	"ogre/internal/pusher"

	"ogre/internal/proxy/httpproxy"

	b "github.com/city-mobil/gobuns/barber"
	"github.com/city-mobil/gobuns/promlib"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	logFileName = "/var/log/ogre.log"
)

var (
	gitCommit     string
	buildDateTime string
	versionTag    string
	buildAuthor   string
)

type infoResponseStruct struct {
	Status        string `json:"status"`
	GitCommit     string `json:"commit"`
	Version       string `json:"version"`
	BuildDateTime string `json:"datetime"`
	BuildAuthor   string `json:"author"`
}

type errorStats struct {
	FailsCount int `json:"errors"`
}

func handleInfo(response []byte) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Error during check info response", err)
		}
	}
}

func main() {
	version := flag.Bool("version", false, "version info")
	pprofPort := flag.String("pprof", ":7001", "use with debug mode")
	debug := flag.Bool("debug", false, "debug mode")
	loggerOutput := flag.String("logger_output", "file", "Use File or Stdout for output")
	pusherConfigCb := pusher.NewConfig()
	httpProxyConfigCb := httpproxy.NewConfig()
	flag.Parse()

	pusherConfig := pusherConfigCb()
	httpProxyConfig := httpProxyConfigCb()

	if *version {
		fmt.Printf("gitCommit: %v\n", gitCommit)
		fmt.Printf("buildDateTime: %v\n", buildDateTime)
		fmt.Printf("buildAuthor: %v\n", buildAuthor)
		os.Exit(0)
	}

	if pusherConfig.SentryHost == "" {
		fmt.Println("Please set sentry_host.\nFor help run 'ogre -h'.")
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if *loggerOutput == "file" {
		logFile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
		if err != nil {
			log.Println("Failed to create log file:", err)
		} else {
			log.SetOutput(logFile)
			go logRotate(ctx, logFile)
		}
	}
	log.SetFlags(log.LstdFlags)

	promlib.SetGlobalNamespace("ogre")
	promlib.IncCntWithLabels("info", promlib.Labels{"version": versionTag})

	psher := pusher.NewPusher(pusherConfig)

	http.HandleFunc("/_status", handleInfo(infoResponse()))
	http.Handle("/metrics", promhttp.Handler())

	if *pprofPort != "" && *debug {
		pprofServer := &http.Server{
			Addr:    *pprofPort,
			Handler: http.DefaultServeMux,
		}

		go func() {
			log.Println("starting pprof", *pprofPort)
			err := pprofServer.ListenAndServe()
			if err != http.ErrServerClosed {
				log.Fatal("failed to start pprof server", err)
			}
		}()
	}

	threshold := uint32(httpProxyConfig.ErrorThreshold / time.Second)
	barber := b.NewBarber([]int{0}, &b.Config{
		Threshold: threshold,
		MaxFails:  50,
	})
	http.HandleFunc("/_errors", handleStats(barber))

	httpProxy := httpproxy.New(httpProxyConfig, psher, barber)
	log.Println("Start server at", httpProxyConfig.Port)
	log.Fatalln(<-httpProxy.Start()) //nolint:gocritic
}

func infoResponse() []byte {
	infoResponseData := infoResponseStruct{
		Status:        "OK",
		GitCommit:     gitCommit,
		BuildDateTime: buildDateTime,
		Version:       versionTag,
		BuildAuthor:   buildAuthor,
	}
	bytes, _ := json.Marshal(infoResponseData)
	return bytes
}

func handleStats(brb b.Barber) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		s := brb.Stats()
		stats := errorStats{
			FailsCount: s.Hosts[len(s.Hosts)-1].FailsCount,
		}
		bytes, _ := json.Marshal(stats)
		w.Header().Add("Content-type", "application/json")
		_, err := w.Write(bytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Error during error stats response ", err.Error())
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

func logRotate(ctx context.Context, rotateFile io.Closer) {
	if rotateFile == nil {
		return
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)
	defer rotateFile.Close()
	for {
		select {
		case <-sigs:
			newLogFile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
			rotateFile.Close()
			if err != nil {
				log.Println("Failed to create log file:", err)
			} else {
				rotateFile = newLogFile
				log.SetOutput(newLogFile)
			}
		case <-ctx.Done():
			return
		}
	}
}
