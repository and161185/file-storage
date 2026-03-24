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
	log.Info("config", "config", cfg)
}
