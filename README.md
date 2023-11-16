# RCN Merlin Exporter

This service monitors RCN's `http://ma.speedtest.rcn.net` endpoint and exports metrics for Prometheus. An image is available at `ghcr.io/kleinpa/rcn-merlin-exporter`.

To try it just run this command and check out the exported metrics on `http://localhost:8080/metrics`:

```
bazel run //cmd/merlin_exporter
```
