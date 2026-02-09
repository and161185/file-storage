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
		"host", cfg.App.Host,
		"port", cfg.App.Port,
		"sizelimit", cfg.App.SizeLimit,
		"timeout", cfg.App.Timeout,
		"image_ext", cfg.Image.Ext,
		"image_max_dimension", cfg.Image.MaxDimension,
		"storage", cfg.App.Storage,
		"file_system_storage_path", cfg.Storage.FileSystem.Path,
		"file_system_lock_lifetime", cfg.Storage.FileSystem.LockLifetime,
	)
}
