package main

import (
	"file-storage/internal/config"
	"log/slog"
	"time"

	"github.com/spf13/pflag"
)

func getConfig() (*config.Config, error) {

	configPathFlag := pflag.String("config", "", "config file path")
	pflag.Int("port", 0, "application port")
	pflag.String("host", "", "application host")
	pflag.String("loglevel", "info", "log level")
	pflag.String("logtype", "json", "log type")
	pflag.String("readtoken", "", "read token")
	pflag.String("writetoken", "", "write token")
	pflag.Duration("timeout", 5*time.Second, "request timeout")
	pflag.Int("sizelimit", 0, "sizelimit")
	pflag.String("imageext", "", "stored image format")
	pflag.Int("imagemaxdimention", 0, "max stored image dimention")
	pflag.String("storage", "", "storage")
	pflag.String("fsstoragepath", "", "file system storage path")
	pflag.Duration("fsstoragelocklifetime", 5*time.Second, "file system lock lifetime")
	pflag.Parse()

	configPath := *configPathFlag

	cfg, err := config.NewConfig(configPath)

	return cfg, err
}

func logConfig(log *slog.Logger, cfg *config.Config) {
	log.Info("config",
		"host", cfg.App.Host,
		"port", cfg.App.Port,
		"sizelimit", cfg.App.SizeLimit,
		"timeout", cfg.App.Timeout,
		"image ext", cfg.Image.Ext,
		"image max dimention", cfg.Image.MaxDimention,
		"storage", cfg.App.Storage,
		"file system storage path", cfg.Storage.FileSystem.Path,
		"file system lock lifetime", cfg.Storage.FileSystem.LockLifetime,
	)
}
