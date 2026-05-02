# Operations

## Purpose

This document describes how to build, run and operate the file storage service.

The service is designed to run as a single-node application with local filesystem storage.

---

## Build and run

Build binary:

```bash
go build -o bin/filestorage ./cmd
```

Run service:

```bash
./bin/filestorage --config configs/config.yaml
```

The service runs as a single process and does not require external dependencies.

---

## Configuration

Configuration is loaded from three sources (in order of priority):

- command-line flags
- environment variables
- configuration file

Example:

```bash
./bin/filestorage --config configs/config.yaml
```

Key configuration areas:

- server settings (host, port, timeouts)
- security (read/write tokens)
- storage (filesystem path, garbage collector settings)
- limits (request size, rate limiting, concurrency)
- image processing settings

Configuration is validated on startup. The service will not start with invalid configuration.

---

## Running with Docker

The service can be built and run as a container.

Build image:

```bash
docker build -t file-storage .
```

Run container:

```bash
docker run -p 8080:8080 \
  -v $(pwd)/configs:/configs \
  file-storage --config /configs/config.yaml
```

The container image is based on `scratch` and contains only the compiled binary.

---

## Shutdown behavior

The service supports graceful shutdown.

On receiving a termination signal:

new requests are not accepted
in-flight requests are allowed to complete
the HTTP server is shut down with a timeout

The shutdown timeout is defined in code.

---

## Garbage collector

The filesystem storage includes a background garbage collector.

Responsibilities:

- remove obsolete file versions
- remove incomplete files left after interrupted writes
- recover version state if the version file is corrupted or missing

Behavior:

- scans the entire storage tree
- processes files per ID
- uses the same per-ID lock as HTTP operations

Garbage collection runs periodically based on configuration.

---

## Logging

The service uses structured logging.

Logs include:

- request metadata (method, path, status, duration)
- request ID for correlation
- errors and panics

Log format and level are configurable.

---

## Metrics

The service exposes Prometheus metrics.

HTTP metrics:

- total requests
- request duration (latency)
- in-flight requests

Storage metrics:

- operation counts
- operation duration
- bytes read and written

Metrics can be used to monitor:

- latency (p50, p95, p99)
- error rates
- throughput
- storage load

---

## Operational limitations
- single-node storage (no distributed coordination)
- no replication or redundancy
- data is stored on local filesystem
- storage cleanup is eventually consistent (via garbage collector)

The service is designed for controlled internal environments.

---