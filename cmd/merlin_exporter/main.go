package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

    "github.com/kleinpa/rcn-merlin-exporter"
)

var httpAddr = flag.String("listen_address", "0.0.0.0:8080", "The address to listen on for HTTP requests.")
var merlinBaseURL = flag.String("merlin_base_url", `http://ma.speedtest.rcn.net`, "Base URL of merlin service")

func main() {
	flag.Parse()

	m := merlin.Client{BaseURL: *merlinBaseURL}

	reg := prometheus.NewRegistry()
	reg.MustRegister(merlin.NewCollector(&m))

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.ListenAndServe(*httpAddr, nil)
}
