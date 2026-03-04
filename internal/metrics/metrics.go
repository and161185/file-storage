package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var HTTPrequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{Name: "http_requests_total", Help: "Total number of http requests"},
	[]string{"operation", "statusCode"},
)

var HTTPrequestsDurationSeconds = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{Name: "http_request_duration_seconds", Help: "Duration of requests in seconds"},
	[]string{"operation"},
)

var HTTPrequestsInFlight = prometheus.NewGauge(
	prometheus.GaugeOpts{Name: "http_in_flight_requests", Help: "In flight requests"},
)

var StorageOperationsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{Name: "storage_operations_total", Help: "Total number of storage operations"},
	[]string{"operation", "result"},
)

var StorageOperationsDurationSeconds = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{Name: "storage_operation_duration_seconds", Help: "Duration of storage operations in seconds"},
	[]string{"operation"},
)

var FileBytesWrittenTotal = prometheus.NewCounter(
	prometheus.CounterOpts{Name: "file_bytes_written_total", Help: "Total number of bytes written"},
)

var FileBytesReadTotal = prometheus.NewCounter(
	prometheus.CounterOpts{Name: "file_bytes_read_total", Help: "Total number of bytes read"},
)

func init() {
	prometheus.MustRegister(HTTPrequestsTotal)
	prometheus.MustRegister(HTTPrequestsDurationSeconds)
	prometheus.MustRegister(HTTPrequestsInFlight)
	prometheus.MustRegister(StorageOperationsTotal)
	prometheus.MustRegister(StorageOperationsDurationSeconds)
	prometheus.MustRegister(FileBytesWrittenTotal)
	prometheus.MustRegister(FileBytesReadTotal)
}
