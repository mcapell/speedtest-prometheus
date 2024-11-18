# Speedtest Prometheus Exporter

This application performs an internet speed test using the `speedtest-go` library and exports the results as Prometheus metrics.

## Features

- Measures latency, download, and upload speeds.
- Pushes metrics to a Prometheus Pushgateway.

## Prerequisites

- Go 1.16 or later
- Access to a Prometheus Pushgateway

## Configuration

Set the `PROMETHEUS_HOST` environment variable to point to your Prometheus Pushgateway:

```bash
export PROMETHEUS_HOST=http://your-prometheus-pushgateway:9091
```

## Usage

Run the application:

```bash
go run main.go
```
