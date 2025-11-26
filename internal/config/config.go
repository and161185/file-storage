package config

import (
	"encoding/json"
	"file-storage/internal/errs"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

const (
	LogTypeJSON = "json"
	LogTypeText = "text"
)

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

type App struct {
	Port int `json:"port"`
}

type Log struct {
	Level string `json:"level"`
	Type  string `json:"type"`
}

type Config struct {
	App App `json:"app"`
	Log Log `json:"log"`
}

func NewConfig(configPath string) (*Config, error) {

	cfg := defaults()

	err := applyConfigFile(&cfg, configPath)
	if err != nil {
		return nil, err
	}

	err = applyEnv(&cfg)
	if err != nil {
		return nil, err
	}

	err = applyFlags(&cfg)
	if err != nil {
		return nil, err
	}

	normalize(&cfg)
	err = validate(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func defaults() Config {
	cfg := Config{Log: Log{Level: LogLevelInfo, Type: LogTypeJSON}, App: App{}}
	return cfg
}

func applyConfigFile(cfg *Config, configPath string) error {
	if configPath == "" {
		return nil
	}

	_, err := os.Stat(configPath)
	if err != nil {
		return err
	}

	b, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, cfg)
	if err != nil {
		return err
	}

	return nil
}

func applyEnv(cfg *Config) error {

	sAppPort := os.Getenv("FILE_STORAGE_APP_PORT")
	if sAppPort != "" {
		port, err := strconv.Atoi(sAppPort)
		if err != nil {
			return err
		}
		cfg.App.Port = port
	}

	sLogLevel := os.Getenv("FILE_STORAGE_LOG_LEVEL")
	if sLogLevel != "" {
		cfg.Log.Level = sLogLevel
	}

	sLogType := os.Getenv("FILE_STORAGE_LOG_TYPE")
	if sLogType != "" {
		cfg.Log.Type = sLogType
	}

	return nil
}

func applyFlags(cfg *Config) error {

	if !pflag.Parsed() {
		return errs.ErrConfigFlagsNotParsed
	}

	fAppPort := pflag.Lookup("port")
	if fAppPort != nil && fAppPort.Changed {
		raw := fAppPort.Value.String()
		port, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		cfg.App.Port = port
	}

	fLogLevel := pflag.Lookup("loglevel")
	if fLogLevel != nil && fLogLevel.Changed {
		cfg.Log.Level = fLogLevel.Value.String()
	}

	fLogType := pflag.Lookup("logtype")
	if fLogType != nil && fLogType.Changed {
		cfg.Log.Type = fLogType.Value.String()
	}

	return nil
}

func normalize(cfg *Config) {
	cfg.Log.Type = strings.ToLower(cfg.Log.Type)
	cfg.Log.Level = strings.ToLower(cfg.Log.Level)
}

func validate(cfg *Config) error {
	if cfg.App.Port < 1 || cfg.App.Port > 65535 {
		return errs.ErrConfigPortOutOfRange
	}

	if cfg.Log.Type != LogTypeJSON && cfg.Log.Type != LogTypeText {
		return errs.ErrConfigWrongLogType
	}

	if cfg.Log.Level != LogLevelDebug && cfg.Log.Level != LogLevelInfo && cfg.Log.Level != LogLevelWarn && cfg.Log.Level != LogLevelError {
		return errs.ErrConfigWrongLogLevel
	}

	return nil
}
