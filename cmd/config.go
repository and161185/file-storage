package main

import (
	"file-storage/internal/config"
	"time"

	"github.com/spf13/pflag"
)

func getConfig() (*config.Config, error) {

	const defaultSizeLimit = 10 * 1024 * 1024

	configPathFlag := pflag.String("config", "", "config file path")
	pflag.String("loglevel", "info", "log level")
	pflag.String("logtype", "json", "log type")
	pflag.Int("port", 0, "application port")
	pflag.String("readtoken", "", "read token")
	pflag.String("writetoken", "", "write token")
	pflag.Duration("timeout", 5*time.Second, "request timeout")
	pflag.Int("sizelimit", defaultSizeLimit, "sizelimit")
	pflag.String("imageext", "jpeg", "stored image format")
	pflag.Int("imagemaxdimention", 2000, "max stored image dimention")
	pflag.Parse()

	configPath := *configPathFlag

	cfg, err := config.NewConfig(configPath)

	return cfg, err
}
