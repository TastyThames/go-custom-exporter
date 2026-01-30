# go-custom-exporter


A lightweight Prometheus exporter written in Go for learning how exporters work
from first principles, without relying on node_exporter or gopsutil.

The project focuses on clarity, simplicity, and production-like structure.

## Features
- Custom /metrics endpoint
- Health check endpoint
- Modular metric collectors
- Designed for production-like infra labs

## Run (dev)
```bash
go run ./cmd/exporter