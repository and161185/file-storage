package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var HTTPrequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{Name: "fs_http_requests_total", Help: "Total number of http requests"},
	[]string{"operation", "statusCode"},
)

var HTTPrequestsDurationSeconds = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{Name: "fs_http_request_duration_seconds", Help: "Duration of requests in seconds"},
	[]string{"operation"},
)

var HTTPrequestsInFlight = prometheus.NewGauge(
	prometheus.GaugeOpts{Name: "fs_http_in_flight_requests", Help: "In flight requests"},
)

var StorageOperationsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{Name: "fs_storage_operations_total", Help: "Total number of storage operations"},
	[]string{"operation", "result"},
)

var StorageOperationsDurationSeconds = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{Name: "fs_storage_operation_duration_seconds", Help: "Duration of storage operations in seconds"},
	[]string{"operation"},
)

var FileBytesWrittenTotal = prometheus.NewCounter(
	prometheus.CounterOpts{Name: "fs_file_bytes_written_total", Help: "Total number of bytes written"},
)

var FileBytesReadTotal = prometheus.NewCounter(
	prometheus.CounterOpts{Name: "fs_file_bytes_read_total", Help: "Total number of bytes read"},
)

var GcRunsTotal = prometheus.NewCounter(
	prometheus.CounterOpts{Name: "fs_gc_runs_total", Help: "Total number of fileservice gc runs"},
)

var GcDurationSeconds = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "fs_gc_duration_seconds",
		Help:    "Duration of gc in seconds",
		Buckets: []float64{10, 30, 60, 300},
	},
)

var GcFilesDeletedTotal = prometheus.NewCounter(
	prometheus.CounterOpts{Name: "fs_gc_files_deleted_total", Help: "Total number of deleted files by gc"},
)

var GcErrorsTotal = prometheus.NewCounter(
	prometheus.CounterOpts{Name: "fs_gc_errors_total", Help: "Total number of gc errors"},
)

var GcRecoveryTotal = prometheus.NewCounter(
	prometheus.CounterOpts{Name: "fs_gc_recovery_total", Help: "Total number of gc recovery"},
)

func init() {
	prometheus.MustRegister(HTTPrequestsTotal)
	prometheus.MustRegister(HTTPrequestsDurationSeconds)
	prometheus.MustRegister(HTTPrequestsInFlight)
	prometheus.MustRegister(StorageOperationsTotal)
	prometheus.MustRegister(StorageOperationsDurationSeconds)
	prometheus.MustRegister(FileBytesWrittenTotal)
	prometheus.MustRegister(FileBytesReadTotal)
	prometheus.MustRegister(GcRunsTotal)
	prometheus.MustRegister(GcDurationSeconds)
	prometheus.MustRegister(GcFilesDeletedTotal)
	prometheus.MustRegister(GcErrorsTotal)
	prometheus.MustRegister(GcRecoveryTotal)
}
