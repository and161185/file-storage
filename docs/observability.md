# Observability

## Purpose

This document describes how the service exposes runtime metrics and how they can be used
to monitor system health and performance.

The service provides metrics for HTTP requests and storage operations.

---

## Metrics

The service exposes Prometheus-compatible metrics.

### HTTP metrics

- total number of requests
- request duration (latency)
- number of in-flight requests

These metrics allow tracking request load, latency and concurrency.

---

### Storage metrics

- total number of storage operations
- storage operation duration
- total bytes read and written

These metrics reflect storage workload and I/O activity.

---

## Interpreting metrics

### Latency

Request latency should be evaluated using percentiles:

- p50 — typical request latency
- p95 — high latency under load
- p99 — worst-case latency

A healthy system usually has:

- p95 close to p50
- p99 not significantly higher than p95

If p95 approaches p99, the system is experiencing instability or resource contention.

---

### Error rate

An increase in error rate may indicate:

- invalid input data
- storage errors
- concurrency limits being reached

Error rate should be monitored together with latency.

---

### Throughput

Throughput (requests per second) should be evaluated together with:

- latency
- CPU usage
- disk activity

Increasing throughput with stable latency indicates healthy scaling.

---

### In-flight requests

In-flight requests indicate current load.

If this value approaches the configured concurrency limit, new requests may be rejected.

---

## What to watch

The following signals are the most important:

- latency (p50, p95, p99)
- error rate
- in-flight requests
- storage operation duration

These signals correspond to standard system health indicators:

- latency
- traffic
- errors
- saturation

---

## Operational interpretation

Typical scenarios:

### Increasing latency

Possible causes:

- high disk I/O
- lock contention (per-ID locking)
- high concurrency

---

### Increased error rate

Possible causes:

- invalid client input
- request limits exceeded
- internal storage errors

---

### High in-flight requests

Indicates that:

- the system is under load
- concurrency limit may be reached soon

---

### High storage latency

Possible causes:

- slow filesystem
- disk saturation
- large number of concurrent writes

---

## Limitations

- metrics are local to a single node
- no distributed aggregation is provided
- no built-in alerting system

Metrics are intended to be consumed by external monitoring systems.