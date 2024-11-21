package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/showwin/speedtest-go/speedtest"
	"go.opentelemetry.io/otel"
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

func main() {
	prometheusHost := os.Getenv("PROMETHEUS_HOST")
	if prometheusHost == "" {
		slog.Error("`PROMETHEUS_HOST` is not defined")
		os.Exit(1)
	}

	if err := initOpentelemetry(); err != nil {
		slog.Error("opentelemetry init", "error", err)
		os.Exit(1)
	}

	speedTest, err := runSpeedTest()
	if err != nil {
		slog.Error("speed test failed", "error", err)
		os.Exit(1)
	}

	if err := pushMetrics(prometheusHost, speedTest); err != nil {
		slog.Error("metrics storage failed", "error", err)
		os.Exit(1)
	}
}

func runSpeedTest() (*speedtest.Server, error) {
	_, span := tracer.Start(context.Background(), "runSpeedTest")
	defer span.End()

	var speedtestClient = speedtest.New()

	serverList, err := speedtestClient.FetchServers()
	if err != nil {
		return nil, fmt.Errorf("error fetching server list: %w", err)
	}

	targets, err := serverList.FindServer(nil)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("server not found")
	}

	target := targets[0]

	slog.Info("start speed test")

	if err := target.PingTest(nil); err != nil {
		return nil, fmt.Errorf("error running the ping test: %w", err)

	}
	if err := target.DownloadTest(); err != nil {
		return nil, fmt.Errorf("error running download test: %w", err)
	}

	if err := target.UploadTest(); err != nil {
		return nil, fmt.Errorf("error running upload test: %w", err)
	}

	slog.Info(fmt.Sprintf("Latency: %s, Download: %s, Upload: %s\n", target.Latency, target.DLSpeed, target.ULSpeed))

	return target, nil
}

func pushMetrics(prometheusHost string, speedTest *speedtest.Server) error {
	_, span := tracer.Start(context.Background(), "pushMetrics")
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

func initOpentelemetry() error {
	ctx := context.Background()

	exp, err := newExporter(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize otel exporter: %w", err)
	}

	tp, err := newTraceProvider(exp, "speedtest")
	if err != nil {
		return fmt.Errorf("failed to initialize otel provider: %w", err)
	}

	// Handle shutdown properly so nothing leaks.
	defer func() { _ = tp.Shutdown(ctx) }()

	otel.SetTracerProvider(tp)
	tracer = tp.Tracer("github.com/mcapell/speedtest-prometheus")

	return nil
}
