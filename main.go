package main

import (
	"context"
	"fmt"
	"os"

	"github.com/showwin/speedtest-go/speedtest"
)

func main() {
	logger := initLogger()
	ctx := WithContext(context.Background(), logger)

	prometheusHost := os.Getenv("PROMETHEUS_HOST")
	if prometheusHost == "" {
		logger.Error("`PROMETHEUS_HOST` is not defined")
		os.Exit(1)
	}

	shutdown, err := initTracer(ctx, "speedtest")
	if err != nil {
		logger.Error("open-telemetry setup", "error", err)
		os.Exit(1)
	}
	defer shutdown()

	ctx, span := tracer.Start(ctx, "speedtest")
	defer span.End()

	speedTest, err := runSpeedTest(ctx)
	if err != nil {
		logger.Error("speed test failed", "error", err)
		os.Exit(1)
	}

	if err := pushMetrics(ctx, prometheusHost, speedTest); err != nil {
		logger.Error("metrics storage failed", "error", err)
		os.Exit(1)
	}
}

func runSpeedTest(ctx context.Context) (*speedtest.Server, error) {
	ctx, span := tracer.Start(ctx, "runSpeedTest")
	defer span.End()

	logger := FromContext(ctx)

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

	logger.Info("start speed test")

	if err := target.PingTest(nil); err != nil {
		return nil, fmt.Errorf("error running the ping test: %w", err)

	}
	if err := target.DownloadTest(); err != nil {
		return nil, fmt.Errorf("error running download test: %w", err)
	}

	if err := target.UploadTest(); err != nil {
		return nil, fmt.Errorf("error running upload test: %w", err)
	}

	logger.Info(fmt.Sprintf("Latency: %s, Download: %s, Upload: %s\n", target.Latency, target.DLSpeed, target.ULSpeed))

	return target, nil
}
