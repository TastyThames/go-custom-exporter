# go-custom-exporter


A lightweight Prometheus exporter written in Go for learning how exporters work
from first principles, without relying on node_exporter or gopsutil.


This project is designed to be **production-like** and suitable for
auto-scaling environments.

## Features

- `/health` endpoint for liveness checks
- `/metrics` endpoint in Prometheus exposition format
- OS metrics collected directly from `/proc`
- Redis Sentinel–aware metrics
- Single static Go binary
- Container-ready (GHCR)

---

## Exported Metrics (partial)

### OS
- `lab_cpu_usage_percent`
- `lab_mem_available_bytes`
- `lab_mem_total_bytes`
- `lab_load1`
- `lab_net_rx_bytes_total`
- `lab_net_tx_bytes_total`
- `lab_uptime_seconds`

### Redis / Sentinel
- `lab_redis_sentinel_up`
- `lab_redis_sentinel_endpoint_up{endpoint}`
- `lab_redis_sentinel_master_info{master_ip,master_port}`
- `lab_redis_master_reachable`
- `lab_redis_local_role`
- `lab_redis_role_mismatch{expected,actual}`

---

## Run locally (dev)

```bash
go run ./cmd/exporter

## Run with Docker
```

```bash
docker run --rm -p 9200:9200 \
  ghcr.io/tastythames/go-custom-exporter:latest
```

## Run with Redis Sentinel 

```bash
docker run -d \
  --restart=always \
  -p 9200:9200 \
  -e REDIS_SENTINELS="ip1:26379,ip2:26379,ip3:26379" \
  -e REDIS_MASTER_NAME="mymaster" \
  ghcr.io/tastythames/go-custom-exporter:latest
```

---

## Prometheus Integration (Auto Scaling friendly)

Use `file_sd_configs` instead of static targets to support scale-out/in.

```bash
scrape_configs:
  - job_name: "web-exporter"
    file_sd_configs:
      - files:
          - /etc/prometheus/web_targets.json
```

---

## Architecture

Exporter runs on **each web instance** and is scraped by a central Prometheus
server. Redis and Sentinel are external dependencies.

```bash
Web VM (Auto Scaling)
├─ go-web container
├─ go-custom-exporter :9200
└─ (optional) node_exporter :9100

obs-1
├─ Prometheus
└─ Grafana
```
