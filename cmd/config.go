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
	cfgCopy := *cfg
	cfgCopy.App.Security.ReadToken = "***"
	cfgCopy.App.Security.WriteToken = "***"
	log.Info("config", "config", cfgCopy)
}
