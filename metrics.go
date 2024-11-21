package main

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/showwin/speedtest-go/speedtest"
)

var (
	latency = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "latency",
		Help:       "Speed test latency.",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	})
	uploadSpeed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "upload_speed",
		Help: "Upload speed in bytes/second.",
	})
	downloadSpeed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "download_speed",
		Help: "Download speed in bytes/second.",
	})
)

func init() {
	prometheus.MustRegister(latency)
	prometheus.MustRegister(uploadSpeed)
	prometheus.MustRegister(downloadSpeed)
}

func pushMetrics(ctx context.Context, prometheusHost string, speedTest *speedtest.Server) error {
	_, span := tracer.Start(ctx, "pushMetrics")
	defer span.End()

	latency.Observe(float64(speedTest.Latency.Microseconds()))
	uploadSpeed.Set(float64(speedTest.ULSpeed))
	downloadSpeed.Set(float64(speedTest.DLSpeed))

	return push.New(prometheusHost, "speedtest").
		Collector(latency).
		Collector(uploadSpeed).
		Collector(downloadSpeed).
		Push()
}
