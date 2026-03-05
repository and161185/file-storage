package metricsstorage

import (
	"context"
	"file-storage/internal/filedata"
	"file-storage/internal/files"
	"file-storage/internal/metrics"
	"io"
	"time"
)

type MetricsStorage struct {
	storage files.Storage
}

func New(storage files.Storage) *MetricsStorage {
	return &MetricsStorage{storage: storage}
}

func (ms *MetricsStorage) Upsert(ctx context.Context, fd *filedata.FileData) (string, error) {

	start := time.Now()

	id, err := ms.storage.Upsert(ctx, fd)

	if fd != nil && err == nil {
		metrics.FileBytesWrittenTotal.Add(float64(len(fd.Data)))
	}

	metrics.StorageOperationsDurationSeconds.WithLabelValues("upsert").Observe(time.Since(start).Seconds())

	metricResult := "ok"
	if err != nil {
		metricResult = "error"
	}
	metrics.StorageOperationsTotal.WithLabelValues("upsert", metricResult).Inc()

	return id, err
}

func (ms *MetricsStorage) Info(ctx context.Context, ID string) (*filedata.FileInfo, error) {
	start := time.Now()

	fd, err := ms.storage.Info(ctx, ID)

	metrics.StorageOperationsDurationSeconds.WithLabelValues("info").Observe(time.Since(start).Seconds())

	metricResult := "ok"
	if err != nil {
		metricResult = "error"
	}
	metrics.StorageOperationsTotal.WithLabelValues("info", metricResult).Inc()

	return fd, err
}

func (ms *MetricsStorage) Content(ctx context.Context, ID string) (*filedata.ContentData, error) {
	start := time.Now()

	fd, err := ms.storage.Content(ctx, ID)

	if err == nil && fd != nil && fd.Data != nil {
		fd.Data = &countingReadCloser{rc: fd.Data,
			onClose: func(n int64) {
				metrics.FileBytesReadTotal.Add(float64(n))
			},
		}
	}

	metrics.StorageOperationsDurationSeconds.WithLabelValues("content").Observe(time.Since(start).Seconds())

	metricResult := "ok"
	if err != nil {
		metricResult = "error"
	}
	metrics.StorageOperationsTotal.WithLabelValues("content", metricResult).Inc()

	return fd, err
}

func (ms *MetricsStorage) Delete(ctx context.Context, ID string) error {
	start := time.Now()

	err := ms.storage.Delete(ctx, ID)

	metrics.StorageOperationsDurationSeconds.WithLabelValues("delete").Observe(time.Since(start).Seconds())

	metricResult := "ok"
	if err != nil {
		metricResult = "error"
	}
	metrics.StorageOperationsTotal.WithLabelValues("delete", metricResult).Inc()

	return err
}

type countingReadCloser struct {
	rc      io.ReadCloser
	n       int64
	onClose func(n int64)
}

func (c *countingReadCloser) Read(p []byte) (int, error) {
	k, err := c.rc.Read(p)
	c.n += int64(k)
	return k, err
}

func (c *countingReadCloser) Close() error {
	err := c.rc.Close()
	if c.onClose != nil {
		c.onClose(c.n)
	}
	return err
}
