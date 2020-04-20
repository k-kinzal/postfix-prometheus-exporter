package main

import (
	"github.com/go-kit/kit/log/level"
	"github.com/k-kinzal/postfix-prometheus-exporter/collector"
	"github.com/k-kinzal/postfix-prometheus-exporter/postfix"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
)

var (
	// Set during go build
	version   string
	gitCommit string

	// Command-line flags
	listenAddress = kingpin.Flag(
		"web.listen-address",
		"Address on which to expose metrics and web interface.",
	).Default(":9154").String()
	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	disableExporterMetrics = kingpin.Flag(
		"web.disable-exporter-metrics",
		"Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).",
	).Bool()
	postfixShowqPath = kingpin.Flag(
		"postfix.showq-path",
		"Path to showq in postfix.",
	).Default("/var/spool/postfix/public/showq").String()
	postfixCollectIntervalSeconds = kingpin.Flag(
		"postfix.interval",
		"Postfix queue in the background to collect statistics on the interval (seconds).",
	).Default("60").Uint64()
)

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting postfix exporter", "version", version, "git commit", gitCommit)

	queue := postfix.NewPostQueue(&postfix.PostQueueOpt{ShowqPath: *postfixShowqPath})
	scheduler := collector.NewPostfixQueueCollectScheduler(queue, logger)
	go func() {
		ch := scheduler.Start(*postfixCollectIntervalSeconds)
		scheduler.Collect()
		<-ch
	}()

	registry := prometheus.NewRegistry()
	if !*disableExporterMetrics {
		registry.MustRegister(
			prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
			prometheus.NewGoCollector(),
		)
	}
	registry.MustRegister(scheduler.Collector())

	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Postfix Exporter</title></head>
			<body>
			<h1>Postfix Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}
