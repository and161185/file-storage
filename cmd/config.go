package main

import (
	"file-storage/internal/config"
	"log/slog"
)

func getConfig(configPathFlag string) (*config.Config, error) {

	cfg, err := config.NewConfig(configPathFlag)

	return cfg, err
}

func logConfig(log *slog.Logger, cfg *config.Config) {
	log.Info("config",
		"host", cfg.App.Server.Host,
		"port", cfg.App.Server.Port,
		"sizelimit", cfg.App.Limits.SizeLimit,
		"handler_timeout", cfg.App.Timeouts.HandlerTimeout,
		"capacity", cfg.App.Limits.RateLimiter.Capacity,
		"refill_rate", cfg.App.Limits.RateLimiter.RefillRate,
		"concurrency_limit", cfg.App.Limits.ConcurrencyLimit,
		"image_ext", cfg.Image.Ext,
		"image_max_dimension", cfg.Image.MaxDimension,
		"storage", cfg.App.Storage,
		"file_system_storage_path", cfg.Storage.FileSystem.Path,
	)
}
