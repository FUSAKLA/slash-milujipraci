package main

import (
	promiddleware "github.com/albertogviana/prometheus-middleware"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"time"
)

var (
	// Version should be set using ldflags during build
	Version = "development"

	debug       = kingpin.Flag("debug", "Enable debug mode.").Bool()
	address     = kingpin.Flag("listen-address", "Address to listen on for API").Default("0.0.0.0:8080").String()
	triggerWord = kingpin.Flag("trigger-word", "Trigger word used").Default("milujipraci").String()
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	kingpin.Version(Version)
	kingpin.Parse()

	cmd := NewCommand(*triggerWord)
	prometheus.MustRegister(cmd)

	middleware := promiddleware.NewPrometheusMiddleware(promiddleware.Opts{Buckets: []float64{0.01, 0.05, 0.1, 0.5}})
	r := mux.NewRouter()
	r.Handle("/metrics", promhttp.Handler())
	r.HandleFunc("/liveness", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("slash-milujipraci is running"))
	})
	r.HandleFunc("/slash-command", cmd.slash).Methods(http.MethodPost, http.MethodGet)
	r.Use(middleware.InstrumentHandlerDuration)

	srv := &http.Server{
		Handler:      handlers.LoggingHandler(os.Stdout, r),
		Addr:         *address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Infof("running on https://%s", *address)
	log.Fatal(srv.ListenAndServe())
}
