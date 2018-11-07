package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
	"github.com/terrycain/389ds_exporter/exporter"
)

var (
	listenPort  = flag.String("web.listen-address", ":9496", "Bind address for prometheus HTTP metrics server")
	metricsPath = flag.String("web.telemetry-path", "/metrics", "Path to expose metrics on")
	ldapAddr    = flag.String("ldap.addr", "localhost:389", "Address of 389ds server")
	ldapUser    = flag.String("ldap.user", "cn=Directory Manager", "389ds Directory Manager user")
	ldapPass    = flag.String("ldap.pass", "", "389ds Directory Manager password")
	ipaDomain   = flag.String("ipa-domain", "", "FreeIPA domain e.g. example.org")
	interval    = flag.Duration("interval", 60*time.Second, "Scrape interval")
	debug       = flag.Bool("debug", false, "Debug logging")
	jsonFormat  = flag.Bool("log-json", false, "JSON formatted log messages")
)

func main() {
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	if *jsonFormat {
		log.SetFormatter(&log.JSONFormatter{})
	}

	if *ldapPass == "" {
		log.Fatal("ldapPass cannot be empty")
	}
	if *ipaDomain == "" {
		log.Fatal("ipaDomain cannot be empty")
	}

	log.Info("Starting prometheus HTTP metrics server on ", *listenPort)
	go StartMetricsServer(*listenPort)

	log.Info("Starting 389ds scraper for ", *ldapAddr)
	for range time.Tick(*interval) {
		log.Debug("Starting metrics scrape")
		exporter.ScrapeMetrics(*ldapAddr, *ldapUser, *ldapPass, *ipaDomain)
	}
}

func StartMetricsServer(bindAddr string) {
	d := http.NewServeMux()
	d.Handle(*metricsPath, promhttp.Handler())
	d.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Consul Exporter</title></head>
             <body>
             <h1>Consul Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </dl>
             <h2>Build</h2>
             <pre>` + version.Info() + ` ` + version.BuildContext() + `</pre>
             </body>
             </html>`))
	})

	err := http.ListenAndServe(bindAddr, d)
	if err != nil {
		log.Fatal("Failed to start metrics server, error is:", err)
	}
}
